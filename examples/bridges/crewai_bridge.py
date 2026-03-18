#!/usr/bin/env python3
"""CrewAI + AiP2P bridge example.

Shows how to use AiP2P as the inter-agent communication layer for CrewAI.
Each CrewAI agent uses AiP2P to discover peers, delegate tasks, and share results
across the P2P network instead of only in-process.

Requirements:
    pip install crewai aip2p
"""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../sdk/python"))

from aip2p import Client

aip2p = Client()


class AiP2PTool:
    """A thin bridge that CrewAI agents can use as a 'tool' to talk over AiP2P."""

    def __init__(self, author: str):
        self.author = author
        aip2p.announce_capability(author=author, tools=["crewai-bridge"])

    def find_agent(self, tool: str) -> str | None:
        agents = aip2p.capabilities(tool=tool)
        return agents[0]["author"] if agents else None

    def assign(self, title: str, body: str, tool: str = "") -> dict | None:
        task = aip2p.task_assign(
            author=self.author,
            title=title,
            body=body,
            extensions={"task.tool": tool} if tool else {},
        )
        return aip2p.wait_task_result(task["infohash"], timeout=120)

    def publish_result(self, title: str, body: str, reply_infohash: str = ""):
        if reply_infohash:
            return aip2p.task_result(
                author=self.author,
                title=title,
                body=body,
                reply_infohash=reply_infohash,
            )
        return aip2p.publish(
            author=self.author,
            kind="task-result",
            title=title,
            body=body,
        )


# --- CrewAI integration ---
# Uncomment below when crewai is installed:
#
# from crewai import Agent, Task, Crew
#
# bridge = AiP2PTool("agent://crewai-coordinator")
#
# researcher = Agent(
#     role="Researcher",
#     goal="Find relevant information",
#     backstory="Expert at gathering data from P2P networks",
#     tools=[bridge],
# )
#
# writer = Agent(
#     role="Writer",
#     goal="Write clear summaries",
#     backstory="Skilled technical writer",
#     tools=[bridge],
# )
#
# research_task = Task(
#     description="Find agents with translate capability on the AiP2P network",
#     agent=researcher,
# )
#
# write_task = Task(
#     description="Summarize the available translation agents",
#     agent=writer,
# )
#
# crew = Crew(agents=[researcher, writer], tasks=[research_task, write_task])
# result = crew.kickoff()
# print(result)


def main():
    """Standalone demo without crewai dependency."""
    print("[crewai-bridge] AiP2P + CrewAI bridge demo")
    print()

    bridge = AiP2PTool("agent://crewai-demo")
    print("[crewai-bridge] bridge initialized")

    # Discover available agents
    all_caps = aip2p.capabilities()
    if all_caps:
        print(f"[crewai-bridge] found {len(all_caps)} agent(s):")
        for cap in all_caps:
            tools = ", ".join(cap.get("tools", []))
            print(f"  {cap['author']}: [{tools}]")
    else:
        print("[crewai-bridge] no agents found on the network")
        print("  hint: run translator.py from task-delegation example first")
        return

    # Try to delegate a task
    translator = bridge.find_agent("translate")
    if translator:
        print(f"[crewai-bridge] delegating to {translator}...")
        result = bridge.assign("Translate", "CrewAI says hello!", tool="translate")
        if result:
            print(f"[crewai-bridge] got: {result.get('body', '')}")
        else:
            print("[crewai-bridge] timeout")


if __name__ == "__main__":
    main()
