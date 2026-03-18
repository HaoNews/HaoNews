#!/usr/bin/env python3
"""Coordinator Agent — discovers capable agents, assigns translation tasks, collects results."""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../sdk/python"))

from aip2p import Client

AUTHOR = "agent://coordinator"

def main():
    client = Client()

    # 1. Announce our capability
    client.announce_capability(
        author=AUTHOR,
        tools=["coordinate", "dispatch"],
        models=[],
        languages=["en", "zh"],
    )
    print("[coordinator] announced capabilities")

    # 2. Discover who can translate
    translators = client.capabilities(tool="translate")
    if not translators:
        print("[coordinator] no translator found, waiting for agents to register...")
        import time
        for _ in range(15):
            time.sleep(2)
            translators = client.capabilities(tool="translate")
            if translators:
                break
    if not translators:
        print("[coordinator] no translator available, exiting")
        return

    print(f"[coordinator] found {len(translators)} translator(s): "
          + ", ".join(t["author"] for t in translators))

    # 3. Assign a translation task
    task = client.task_assign(
        author=AUTHOR,
        title="Translate to Chinese",
        body="Hello, this is a test message for the AiP2P multi-agent demo.",
        channel="tasks",
        priority="high",
        extensions={"task.tool": "translate", "task.target_lang": "zh"},
    )
    infohash = task["infohash"]
    print(f"[coordinator] task assigned: {infohash}")

    # 4. Wait for result
    print("[coordinator] waiting for task result...")
    result = client.wait_task_result(infohash, timeout=120, poll_interval=3)
    if result:
        print(f"[coordinator] got result from {result.get('author', '?')}:")
        print(f"  {result.get('body', '')}")
    else:
        print("[coordinator] timeout waiting for result")


if __name__ == "__main__":
    main()
