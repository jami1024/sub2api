#!/usr/bin/env python3
import importlib.util
import pathlib
import unittest


SCRIPT_PATH = pathlib.Path(__file__).resolve().parent / "tools" / "compare_stream_targets.py"
SPEC = importlib.util.spec_from_file_location("compare_stream_targets", SCRIPT_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC and SPEC.loader
SPEC.loader.exec_module(MODULE)


class CompareStreamTargetsTest(unittest.TestCase):
    def test_summarize_sse_body_extracts_tail_events_and_done(self) -> None:
        body = "\n".join(
            [
                "event: response.created",
                'data: {"type":"response.created","response":{"id":"resp_1"}}',
                "",
                "event: response.output_text.delta",
                'data: {"type":"response.output_text.delta","delta":"Hi"}',
                "",
                "event: response.completed",
                'data: {"type":"response.completed","response":{"id":"resp_1","output":[{"content":[{"text":"Hi"}]}]}}',
                "",
                "data: [DONE]",
                "",
            ]
        )

        summary = MODULE.summarize_stream_text(body)

        self.assertTrue(summary["is_sse_like"])
        self.assertTrue(summary["has_done_marker"])
        self.assertEqual(
            summary["tail_event_types"],
            ["response.created", "response.output_text.delta", "response.completed"],
        )
        self.assertEqual(
            summary["tail_payload_types"],
            ["response.created", "response.output_text.delta", "response.completed"],
        )
        self.assertEqual(summary["last_payload_type"], "response.completed")

    def test_summarize_sse_body_handles_eof_without_done(self) -> None:
        body = "\n".join(
            [
                "event: response.created",
                'data: {"type":"response.created","response":{"id":"resp_2"}}',
                "",
                "event: response.completed",
                'data: {"type":"response.completed","response":{"id":"resp_2"}}',
                "",
                "",
            ]
        )

        summary = MODULE.summarize_stream_text(body)

        self.assertTrue(summary["is_sse_like"])
        self.assertFalse(summary["has_done_marker"])
        self.assertEqual(summary["last_payload_type"], "response.completed")
        self.assertTrue(summary["ends_with_event_boundary"])

    def test_resolve_body_arg_supports_inline_and_file(self) -> None:
        self.assertEqual(
            MODULE.resolve_body_arg(None, '{"provider":"mine"}'),
            ("--data-binary", '{"provider":"mine"}'),
        )
        self.assertEqual(
            MODULE.resolve_body_arg("req.json", None),
            ("--data-binary", "@req.json"),
        )

    def test_load_saved_capture_supports_directory_and_file(self) -> None:
        import json
        import tempfile

        sample = {
            "name": "mine",
            "status_code": 200,
            "response_headers": {"x-request-id": "rid-1"},
            "stream_summary": {"has_done_marker": True, "last_payload_type": "response.completed", "tail_payload_types": []},
            "curl_meta": {"time_total": "1.23", "time_starttransfer": "0.56"},
        }

        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = pathlib.Path(tmp)
            summary_path = tmp_path / "summary.json"
            summary_path.write_text(json.dumps(sample), encoding="utf-8")

            loaded_from_dir = MODULE.load_saved_capture(tmp_path)
            loaded_from_file = MODULE.load_saved_capture(summary_path)

            self.assertEqual(loaded_from_dir["name"], "mine")
            self.assertEqual(loaded_from_file["response_headers"]["x-request-id"], "rid-1")

    def test_compare_saved_captures_reuses_same_diff_schema(self) -> None:
        capture_a = {
            "status_code": 200,
            "response_headers": {"x-request-id": "rid-a", "content-type": "text/event-stream"},
            "stream_summary": {
                "has_done_marker": True,
                "last_payload_type": "response.completed",
                "tail_payload_types": ["response.created", "response.completed"],
            },
            "curl_meta": {"time_starttransfer": "1.0", "time_total": "2.0"},
        }
        capture_b = {
            "status_code": 200,
            "response_headers": {"x-request-id": "rid-b", "content-type": "text/event-stream"},
            "stream_summary": {
                "has_done_marker": False,
                "last_payload_type": "response.output_item.done",
                "tail_payload_types": ["response.created", "response.output_item.done"],
            },
            "curl_meta": {"time_starttransfer": "1.5", "time_total": "2.5"},
        }

        diff = MODULE.compare_results(capture_a, capture_b)

        self.assertEqual(diff["x_request_id_diff"], ["rid-a", "rid-b"])
        self.assertEqual(diff["done_marker_diff"], [True, False])
        self.assertEqual(
            diff["tail_payload_types_diff"],
            [
                ["response.created", "response.completed"],
                ["response.created", "response.output_item.done"],
            ],
        )


if __name__ == "__main__":
    unittest.main()
