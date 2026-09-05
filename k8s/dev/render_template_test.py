import json
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]


class FacadeRoutingTest(unittest.TestCase):
    def render(self, environment, mode=None):
        config = json.loads((ROOT / "tilt_config.json.example").read_text())
        config[environment].pop("api_facade", None)
        if mode is not None:
            config[environment]["api_facade"] = mode
        with tempfile.TemporaryDirectory() as directory:
            config_path = Path(directory) / "config.json"
            output = Path(directory) / "access_rules.yaml"
            config_path.write_text(json.dumps(config))
            result = subprocess.run(
                [
                    sys.executable, str(ROOT / "k8s/dev/render_template.py"),
                    "--config", str(config_path), "--environment", environment,
                    "--src", str(ROOT / "infra/dev/ory/access_rules.yaml"),
                    "--dst", str(output),
                ],
                capture_output=True, text=True, check=False,
            )
            return result, output.read_text() if output.exists() else ""

    def test_facade_changes_only_five_upstreams_and_their_prefix_stripping(self):
        for environment in ("local", "shared"):
            with self.subTest(environment=environment):
                result, direct = self.render(environment)
                self.assertEqual(result.returncode, 0, result.stderr)
                result, disabled = self.render(environment, False)
                self.assertEqual(result.returncode, 0, result.stderr)
                self.assertEqual(disabled, direct)
                result, facade = self.render(environment, True)
                self.assertEqual(result.returncode, 0, result.stderr)

                expected = direct
                for service in ("authz", "content", "immersion", "profile"):
                    upstream = (
                        '    url: "http://{0}-api.tdk-{0}-api:80"\n'
                        '    strip_path: "/api/internal/{0}/"'
                    ).format(service)
                    self.assertEqual(direct.count(upstream), 2 if service == "immersion" else 1)
                    expected = expected.replace(
                        upstream,
                        '    url: "http://tadoku-api.tdk-tadoku-api:80"\n'
                        '    strip_path: ""',
                    )
                self.assertEqual(facade, expected)
                self.assertEqual(facade.count('url: "http://tadoku-api.tdk-tadoku-api:80"'), 5)
                self.assertNotIn("{{TADOKU_", facade)

    def test_facade_requires_a_json_boolean(self):
        result, output = self.render("local", "false")
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('"api_facade" must be true or false', result.stderr)
        self.assertEqual(output, "")


if __name__ == "__main__":
    unittest.main()
