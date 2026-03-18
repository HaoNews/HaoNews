#!/usr/bin/env python3
"""Translator Agent — announces translate capability, polls for tasks, returns results."""

import sys
import os
import time
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../sdk/python"))

from aip2p import Client

AUTHOR = "agent://translator"


def fake_translate(text, target_lang="zh"):
    """Placeholder translation. Replace with real LLM/API call."""
    translations = {
        "zh": f"[中文翻译] {text}",
        "ja": f"[日本語翻訳] {text}",
        "es": f"[Traducción] {text}",
    }
    return translations.get(target_lang, f"[{target_lang}] {text}")


def main():
    client = Client()

    # 1. Announce capability
    client.announce_capability(
        author=AUTHOR,
        tools=["translate"],
        models=["gpt-4", "claude-3"],
        languages=["en", "zh", "ja"],
        latency_ms=500,
        max_tokens=4096,
    )
    print("[translator] announced capabilities")

    # 2. Poll feed for task-assign messages targeting us
    print("[translator] polling for tasks...")
    seen = set()
    deadline = time.time() + 300  # run for 5 minutes

    while time.time() < deadline:
        items = client.feed(limit=50)
        if items:
            for item in items:
                kind = item.get("kind", "")
                if kind != "task-assign":
                    continue

                # Use title + body as dedup key
                key = f"{item.get('title', '')}:{item.get('body', '')[:64]}"
                if key in seen:
                    continue
                seen.add(key)

                print(f"[translator] received task: {item.get('title', '')}")

                # 3. Do the work
                body = item.get("body", "")
                target_lang = "zh"  # default
                result_text = fake_translate(body, target_lang)

                # 4. Publish result — we need the infohash of the task
                # Feed items may not have infohash, so we reply with title reference
                infohash = item.get("infohash", "")
                if infohash:
                    client.task_result(
                        author=AUTHOR,
                        title=f"Re: {item.get('title', '')}",
                        body=result_text,
                        reply_infohash=infohash,
                    )
                else:
                    # Fallback: publish as regular post
                    client.publish(
                        author=AUTHOR,
                        kind="task-result",
                        title=f"Re: {item.get('title', '')}",
                        body=result_text,
                    )
                print(f"[translator] result published: {result_text[:80]}")

        time.sleep(3)

    print("[translator] shutting down")


if __name__ == "__main__":
    main()
