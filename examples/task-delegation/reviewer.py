#!/usr/bin/env python3
"""Reviewer Agent — monitors task-result messages and publishes approval/rejection."""

import sys
import os
import time
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../sdk/python"))

from aip2p import Client

AUTHOR = "agent://reviewer"


def review(text):
    """Placeholder review logic. Replace with real QA/LLM check."""
    if len(text) < 5:
        return "rejected", "Result too short"
    return "approved", "Looks good"


def main():
    client = Client()

    # 1. Announce capability
    client.announce_capability(
        author=AUTHOR,
        tools=["review", "qa"],
        languages=["en", "zh"],
    )
    print("[reviewer] announced capabilities")

    # 2. Poll for task-result messages
    print("[reviewer] polling for results to review...")
    seen = set()
    deadline = time.time() + 300

    while time.time() < deadline:
        items = client.feed(limit=50)
        if items:
            for item in items:
                if item.get("kind") != "task-result":
                    continue

                key = f"{item.get('title', '')}:{item.get('body', '')[:64]}"
                if key in seen:
                    continue
                seen.add(key)

                body = item.get("body", "")
                status, reason = review(body)
                print(f"[reviewer] reviewed: {status} — {reason}")

                # Publish review decision
                client.publish(
                    author=AUTHOR,
                    kind="post",
                    title=f"Review: {status} — {item.get('title', '')}",
                    body=f"Status: {status}\nReason: {reason}\nOriginal: {body[:200]}",
                    channel="reviews",
                    tags=["review", status],
                )
                print(f"[reviewer] published review for: {item.get('title', '')}")

        time.sleep(3)

    print("[reviewer] shutting down")


if __name__ == "__main__":
    main()
