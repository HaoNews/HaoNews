#!/usr/bin/env python3
"""LangGraph + AiP2P bridge example.

Shows how to use AiP2P as the communication layer in a LangGraph workflow.
A coordinator node discovers agents via capability API, delegates tasks,
and collects results — all through the AiP2P local HTTP API.

Requirements:
    pip install langgraph aip2p
"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../sdk/python"))

from aip2p import Client

# --- AiP2P bridge functions ---

aip2p = Client()
AUTHOR = "agent://langgraph-coordinator"


def discover_agent(tool: str) -> str | None:
    """Find an agent with a specific tool capability."""
    agents = aip2p.capabilities(tool=tool)
    if agents:
        return agents[0]["author"]
    return None


def delegate_task(title: str, body: str, tool: str) -> dict | None:
    """Assign a task via AiP2P and wait for the result."""
    task = aip2p.task_assign(
        author=AUTHOR,
        title=title,
        body=body,
        extensions={"task.tool": tool},
    )
    result = aip2p.wait_task_result(task["infohash"], timeout=120)
    return result


# --- LangGraph workflow ---
# Uncomment below when langgraph is installed:
#
# from langgraph.graph import StateGraph
# from typing import TypedDict
#
# class State(TypedDict):
#     text: str
#     translated: str
#     reviewed: bool
#
# def translate_node(state: State) -> State:
#     agent = discover_agent("translate")
#     if not agent:
#         return {**state, "translated": "[no translator available]"}
#     result = delegate_task("Translate", state["text"], "translate")
#     return {**state, "translated": result["body"] if result else ""}
#
# def review_node(state: State) -> State:
#     agent = discover_agent("review")
#     if not agent:
#         return {**state, "reviewed": True}  # auto-approve if no reviewer
#     result = delegate_task("Review", state["translated"], "review")
#     approved = result and "approved" in result.get("body", "").lower()
#     return {**state, "reviewed": approved}
#
# graph = StateGraph(State)
# graph.add_node("translate", translate_node)
# graph.add_node("review", review_node)
# graph.add_edge("translate", "review")
# graph.set_entry_point("translate")
# graph.set_finish_point("review")
# app = graph.compile()
#
# result = app.invoke({"text": "Hello world", "translated": "", "reviewed": False})
# print(result)


def main():
    """Standalone demo without langgraph dependency."""
    print("[langgraph-bridge] AiP2P + LangGraph bridge demo")
    print()

    # Announce coordinator
    aip2p.announce_capability(author=AUTHOR, tools=["coordinate"])
    print("[langgraph-bridge] announced coordinator capability")

    # Discover translator
    translator = discover_agent("translate")
    if translator:
        print(f"[langgraph-bridge] found translator: {translator}")
        result = delegate_task(
            "Translate to Chinese",
            "Hello, this is a LangGraph bridge demo.",
            "translate",
        )
        if result:
            print(f"[langgraph-bridge] result: {result.get('body', '')}")
        else:
            print("[langgraph-bridge] no result (timeout)")
    else:
        print("[langgraph-bridge] no translator found")
        print("  hint: run translator.py from task-delegation example first")


if __name__ == "__main__":
    main()
