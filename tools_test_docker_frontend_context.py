from pathlib import Path
import unittest


ROOT = Path(__file__).resolve().parent


class DockerFrontendContextTest(unittest.TestCase):
    def test_docker_frontend_builder_includes_legal_markdown_sources(self) -> None:
        dockerfile = (ROOT / "Dockerfile").read_text()
        dockerignore = (ROOT / ".dockerignore").read_text()

        self.assertIn("docs/legal/admin-compliance.zh.md", dockerfile)
        self.assertIn("docs/legal/admin-compliance.en.md", dockerfile)
        self.assertIn("!docs/legal/admin-compliance.zh.md", dockerignore)
        self.assertIn("!docs/legal/admin-compliance.en.md", dockerignore)


if __name__ == "__main__":
    unittest.main()
