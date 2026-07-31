#!/usr/bin/env python3
"""
对同一份请求分别打两个目标地址，保存响应头/响应体/时序信息，并比较 SSE 尾部行为。

典型用途：
  - 比较自己的 sub2api 与别人的 sub2api 在 /v1/responses 流式收尾上的差异
  - 判断是否存在 response.completed / [DONE] / EOF 收尾差异
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import tempfile
from datetime import datetime
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Capture and compare raw streaming responses.")
    subparsers = parser.add_subparsers(dest="command", required=True)

    capture = subparsers.add_parser("capture", help="单次请求并保存响应")
    capture.add_argument("--target-url", required=True, help="目标 URL（例如你的 CCH /v1/responses）")
    capture.add_argument("--method", default="POST", help="HTTP 方法，默认 POST")
    capture.add_argument("--header", action="append", default=[], help="可重复，格式：'Header: value'")
    capture.add_argument("--data-file", help="请求体文件路径")
    capture.add_argument("--data-inline", help="直接传入请求体字符串")
    capture.add_argument("--max-time", type=int, default=120, help="curl 最大执行时间（秒）")
    capture.add_argument("--output-dir", help="输出目录；默认自动创建到 /tmp")
    capture.add_argument("--name", default="capture", help="本次抓取的名字，写入 summary.json")
    capture.add_argument("--insecure", action="store_true", help="curl -k，跳过 TLS 校验")

    compare = subparsers.add_parser("compare", help="读取两次已保存的抓取结果并对比")
    compare.add_argument("--capture-a", required=True, help="第一次抓取结果目录或 summary.json")
    compare.add_argument("--capture-b", required=True, help="第二次抓取结果目录或 summary.json")
    return parser.parse_args()


def extract_json_type(payload: str) -> str | None:
    try:
        obj = json.loads(payload)
    except json.JSONDecodeError:
        return None
    if isinstance(obj, dict):
        raw = obj.get("type")
        if isinstance(raw, str) and raw.strip():
            return raw.strip()
    return None


def summarize_stream_text(text: str) -> dict[str, Any]:
    event_types: list[str] = []
    payload_types: list[str] = []
    data_payloads: list[str] = []
    has_done_marker = False
    has_event_field = False

    lines = text.splitlines()
    for raw_line in lines:
        line = raw_line.strip()
        if not line:
            continue
        if line.startswith("event:"):
            has_event_field = True
            event_name = line[6:].strip()
            if event_name:
                event_types.append(event_name)
            continue
        if not line.startswith("data:"):
            continue

        payload = line[5:].lstrip()
        if payload == "[DONE]":
            has_done_marker = True
            continue
        if payload:
            data_payloads.append(payload)
            payload_type = extract_json_type(payload)
            if payload_type:
                payload_types.append(payload_type)

    tail_event_types = event_types[-10:]
    tail_payload_types = payload_types[-10:]
    trimmed = text.rstrip()

    return {
        "is_sse_like": has_event_field or bool(data_payloads) or has_done_marker,
        "line_count": len(lines),
        "data_event_count": len(data_payloads),
        "has_done_marker": has_done_marker,
        "ends_with_event_boundary": text.endswith("\n\n") or trimmed.endswith("\n\n"),
        "tail_event_types": tail_event_types,
        "tail_payload_types": tail_payload_types,
        "last_event_type": tail_event_types[-1] if tail_event_types else None,
        "last_payload_type": tail_payload_types[-1] if tail_payload_types else None,
        "last_non_empty_line": next((line for line in reversed(lines) if line.strip()), None),
    }


def parse_last_header_block(header_text: str) -> tuple[int | None, dict[str, str]]:
    blocks = [block for block in header_text.split("\r\n\r\n") if block.strip()]
    if not blocks:
        blocks = [block for block in header_text.split("\n\n") if block.strip()]
    if not blocks:
        return None, {}

    block = blocks[-1]
    lines = [line.strip("\r") for line in block.splitlines() if line.strip()]
    if not lines:
        return None, {}

    status_code = None
    parts = lines[0].split()
    if len(parts) >= 2 and parts[1].isdigit():
        status_code = int(parts[1])

    headers: dict[str, str] = {}
    for line in lines[1:]:
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        headers[key.strip().lower()] = value.strip()
    return status_code, headers


def run_probe(
    name: str,
    url: str,
    method: str,
    headers: list[str],
    body_arg: tuple[str, str] | None,
    insecure: bool,
    max_time: int,
    output_dir: Path,
) -> dict[str, Any]:
    header_path = output_dir / f"{name}.headers.txt"
    body_path = output_dir / f"{name}.body.txt"

    meta_format = json.dumps(
        {
            "http_code": "%{http_code}",
            "time_namelookup": "%{time_namelookup}",
            "time_connect": "%{time_connect}",
            "time_appconnect": "%{time_appconnect}",
            "time_starttransfer": "%{time_starttransfer}",
            "time_total": "%{time_total}",
            "size_download": "%{size_download}",
            "content_type": "%{content_type}",
            "remote_ip": "%{remote_ip}",
        }
    )

    cmd = [
        "curl",
        "--silent",
        "--show-error",
        "--no-buffer",
        "--location",
        "--request",
        method,
        "--url",
        url,
        "--dump-header",
        str(header_path),
        "--output",
        str(body_path),
        "--write-out",
        meta_format,
        "--max-time",
        str(max_time),
    ]
    if insecure:
        cmd.append("--insecure")
    for header in headers:
        cmd.extend(["--header", header])
    if body_arg:
        flag, value = body_arg
        cmd.extend([flag, value])

    proc = subprocess.run(cmd, capture_output=True, text=True)
    stdout = proc.stdout.strip()
    stderr = proc.stderr.strip()
    try:
        curl_meta = json.loads(stdout) if stdout else {}
    except json.JSONDecodeError:
        curl_meta = {"raw": stdout}

    header_text = header_path.read_text(encoding="utf-8", errors="replace") if header_path.exists() else ""
    body_text = body_path.read_text(encoding="utf-8", errors="replace") if body_path.exists() else ""
    status_code, response_headers = parse_last_header_block(header_text)

    return {
        "name": name,
        "url": url,
        "exit_code": proc.returncode,
        "stderr": stderr,
        "curl_meta": curl_meta,
        "status_code": status_code,
        "response_headers": response_headers,
        "stream_summary": summarize_stream_text(body_text),
        "artifacts": {
            "headers": str(header_path),
            "body": str(body_path),
        },
    }


def print_summary(result: dict[str, Any]) -> None:
    meta = result["curl_meta"]
    stream = result["stream_summary"]
    headers = result["response_headers"]
    print(f"\n== {result['name']} ==")
    print(f"url: {result['url']}")
    print(f"curl_exit: {result['exit_code']}")
    print(f"http_status: {result['status_code']}")
    print(
        "timing:"
        f" starttransfer={meta.get('time_starttransfer')}"
        f" total={meta.get('time_total')}"
        f" download={meta.get('size_download')}"
    )
    print(f"content-type: {headers.get('content-type')}")
    print(f"x-request-id: {headers.get('x-request-id')}")
    print(f"has_done_marker: {stream['has_done_marker']}")
    print(f"last_payload_type: {stream['last_payload_type']}")
    print(f"tail_payload_types: {stream['tail_payload_types']}")
    print(f"last_non_empty_line: {stream['last_non_empty_line']}")
    if result["stderr"]:
        print(f"stderr: {result['stderr']}")
    print(f"headers_file: {result['artifacts']['headers']}")
    print(f"body_file: {result['artifacts']['body']}")


def compare_results(a: dict[str, Any], b: dict[str, Any]) -> dict[str, Any]:
    return {
        "status_diff": [a.get("status_code"), b.get("status_code")],
        "x_request_id_diff": [
            a["response_headers"].get("x-request-id"),
            b["response_headers"].get("x-request-id"),
        ],
        "content_type_diff": [
            a["response_headers"].get("content-type"),
            b["response_headers"].get("content-type"),
        ],
        "done_marker_diff": [
            a["stream_summary"]["has_done_marker"],
            b["stream_summary"]["has_done_marker"],
        ],
        "last_payload_type_diff": [
            a["stream_summary"]["last_payload_type"],
            b["stream_summary"]["last_payload_type"],
        ],
        "tail_payload_types_diff": [
            a["stream_summary"]["tail_payload_types"],
            b["stream_summary"]["tail_payload_types"],
        ],
        "time_starttransfer_diff": [
            a["curl_meta"].get("time_starttransfer"),
            b["curl_meta"].get("time_starttransfer"),
        ],
        "time_total_diff": [
            a["curl_meta"].get("time_total"),
            b["curl_meta"].get("time_total"),
        ],
    }


def resolve_body_arg(data_file: str | None, data_inline: str | None) -> tuple[str, str] | None:
    if data_file and data_inline is not None:
        raise ValueError("同一个目标不能同时提供 data-file 和 data-inline")
    if data_file:
        return ("--data-binary", f"@{data_file}")
    if data_inline is not None:
        return ("--data-binary", data_inline)
    return None


def load_saved_capture(path_like: str | Path) -> dict[str, Any]:
    path = Path(path_like)
    if path.is_dir():
        path = path / "summary.json"
    return json.loads(path.read_text(encoding="utf-8"))


def write_summary_file(output_dir: Path, summary: dict[str, Any]) -> Path:
    summary_path = output_dir / "summary.json"
    summary_path.write_text(json.dumps(summary, ensure_ascii=False, indent=2), encoding="utf-8")
    return summary_path


def handle_capture(args: argparse.Namespace) -> int:
    if bool(args.data_file) == bool(args.data_inline):
        print("capture 模式必须二选一提供 --data-file 或 --data-inline", file=sys.stderr)
        return 2

    timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
    output_dir = (
        Path(args.output_dir)
        if args.output_dir
        else Path(tempfile.mkdtemp(prefix=f"capture-stream-{timestamp}-"))
    )
    output_dir.mkdir(parents=True, exist_ok=True)

    result = run_probe(
        name=args.name,
        url=args.target_url,
        method=args.method,
        headers=list(args.header),
        body_arg=resolve_body_arg(args.data_file, args.data_inline),
        insecure=args.insecure,
        max_time=args.max_time,
        output_dir=output_dir,
    )
    summary = {
        "generated_at": datetime.now().isoformat(),
        "output_dir": str(output_dir),
        **result,
    }
    summary_path = write_summary_file(output_dir, summary)
    print_summary(result)
    print(f"\nsummary_file: {summary_path}")
    return 0


def handle_compare(args: argparse.Namespace) -> int:
    capture_a = load_saved_capture(args.capture_a)
    capture_b = load_saved_capture(args.capture_b)
    comparison = compare_results(capture_a, capture_b)
    print_summary(capture_a)
    print_summary(capture_b)
    print("\n== comparison ==")
    print(json.dumps(comparison, ensure_ascii=False, indent=2))
    return 0


def main() -> int:
    args = parse_args()
    if args.command == "capture":
        return handle_capture(args)
    if args.command == "compare":
        return handle_compare(args)
    print(f"未知命令: {args.command}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
