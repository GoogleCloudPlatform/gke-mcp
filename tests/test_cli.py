import pytest
from click.testing import CliRunner
from gke_mcp.cli import main

def test_cli_help():
    runner = CliRunner()
    res = runner.invoke(main, ["--help"])
    assert res.exit_code == 0
    assert "An MCP Server for Google Kubernetes Engine" in res.output
    assert "--server-mode" in res.output
    assert "install" in res.output

def test_cli_install_help():
    runner = CliRunner()
    res = runner.invoke(main, ["install", "--help"])
    assert res.exit_code == 0
    assert "gemini-cli" in res.output
    assert "cursor" in res.output
    assert "claude-desktop" in res.output
    assert "claude-code" in res.output

def test_cli_default_run(monkeypatch):
    runner = CliRunner()
    server_called = False

    def mock_start_server(cfg, server_mode, server_host, server_port, origins):
        nonlocal server_called
        server_called = True
        assert server_mode == "stdio"
        assert server_host == "127.0.0.1"
        assert server_port == 8080

    monkeypatch.setattr("gke_mcp.server.start_server", mock_start_server)

    res = runner.invoke(main, [])
    assert res.exit_code == 0
    assert server_called is True
