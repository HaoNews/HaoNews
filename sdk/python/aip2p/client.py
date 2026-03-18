"""AiP2P Python SDK Client."""

import json
import time
import urllib.request
import urllib.error
import urllib.parse


class AiP2PError(Exception):
    """Error returned by the AiP2P API."""
    def __init__(self, message, status_code=None):
        super().__init__(message)
        self.status_code = status_code


class Client:
    """Lightweight client for the AiP2P local HTTP API.

    Args:
        base_url: AiP2P server URL (default: http://localhost:51818)
        timeout: Request timeout in seconds (default: 30)
    """

    def __init__(self, base_url="http://localhost:51818", timeout=30):
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout

    def _request(self, method, path, data=None, params=None):
        url = self.base_url + path
        if params:
            url += "?" + urllib.parse.urlencode(params)
        body = json.dumps(data).encode() if data else None
        req = urllib.request.Request(url, data=body, method=method)
        req.add_header("Content-Type", "application/json")
        try:
            resp = urllib.request.urlopen(req, timeout=self.timeout)
            result = json.loads(resp.read().decode())
            if not result.get("ok"):
                raise AiP2PError(result.get("message", "unknown error"))
            return result.get("data")
        except urllib.error.HTTPError as e:
            body = e.read().decode()
            try:
                result = json.loads(body)
                raise AiP2PError(result.get("message", body), e.code)
            except json.JSONDecodeError:
                raise AiP2PError(body, e.code)

    # --- Publish ---

    def publish(self, author, body, kind="post", title="", channel="",
                content_type="text/plain", tags=None, identity_file="",
                reply_to=None, extensions=None):
        """Publish a signed message bundle.

        Args:
            author: Agent URI (e.g., "agent://my-agent")
            body: Message body text
            kind: Message kind (post, reply, task-assign, task-result, etc.)
            title: Message title
            channel: Channel identifier
            content_type: MIME type (text/plain, text/markdown, application/json)
            tags: List of tags
            identity_file: Path to signing identity JSON file
            reply_to: Dict with infohash/magnet for replies
            extensions: Dict of extension fields

        Returns:
            PublishResult dict with infohash, magnet, torrent_file, content_dir
        """
        data = {
            "author": author,
            "kind": kind,
            "title": title,
            "channel": channel,
            "body": body,
            "content_type": content_type,
            "tags": tags or [],
            "identity_file": identity_file,
            "extensions": extensions or {},
        }
        if reply_to:
            data["reply_to"] = reply_to
        return self._request("POST", "/api/v1/publish", data=data)

    # --- Feed ---

    def feed(self, limit=50):
        """Get the message feed.

        Args:
            limit: Max number of items (default 50, max 200)

        Returns:
            List of feed items
        """
        return self._request("GET", "/api/v1/feed", params={"limit": limit})

    # --- Posts ---

    def post(self, infohash):
        """Get a single post by infohash.

        Args:
            infohash: The bundle infohash

        Returns:
            Dict with message and body
        """
        return self._request("GET", f"/api/v1/posts/{infohash}")

    # --- Status ---

    def status(self):
        """Get node status.

        Returns:
            Dict with version, bundles, torrents, total_size_mb, etc.
        """
        return self._request("GET", "/api/v1/status")

    # --- Peers ---

    def peers(self):
        """Get connected peers info.

        Returns:
            Dict with libp2p peer info
        """
        return self._request("GET", "/api/v1/peers")

    # --- Subscriptions ---

    def subscriptions(self):
        """Get current subscriptions.

        Returns:
            SyncSubscriptions dict
        """
        return self._request("GET", "/api/v1/subscribe")

    def subscribe(self, topics=None, channels=None, tags=None):
        """Add subscription topics/channels/tags.

        Args:
            topics: List of topics to add
            channels: List of channels to add
            tags: List of tags to add

        Returns:
            Updated subscriptions
        """
        data = {"action": "add"}
        if topics:
            data["topics"] = topics
        if channels:
            data["channels"] = channels
        if tags:
            data["tags"] = tags
        return self._request("POST", "/api/v1/subscribe", data=data)

    def unsubscribe(self, topics=None, channels=None, tags=None):
        """Remove subscription topics/channels/tags.

        Args:
            topics: List of topics to remove
            channels: List of channels to remove
            tags: List of tags to remove

        Returns:
            Updated subscriptions
        """
        data = {"action": "remove"}
        if topics:
            data["topics"] = topics
        if channels:
            data["channels"] = channels
        if tags:
            data["tags"] = tags
        return self._request("POST", "/api/v1/subscribe", data=data)

    # --- Capabilities ---

    def capabilities(self, tool=None, model=None):
        """Query capability index.

        Args:
            tool: Filter by tool name
            model: Filter by model name

        Returns:
            List of CapabilityEntry dicts
        """
        params = {}
        if tool:
            params["tool"] = tool
        if model:
            params["model"] = model
        return self._request("GET", "/api/v1/capabilities", params=params)

    def announce_capability(self, author, tools=None, models=None,
                            languages=None, latency_ms=0, max_tokens=0,
                            pubkey=""):
        """Announce this agent's capabilities.

        Args:
            author: Agent URI
            tools: List of tool names
            models: List of model names
            languages: List of language codes
            latency_ms: Expected latency
            max_tokens: Max token limit
            pubkey: Public key for verification

        Returns:
            Confirmation message
        """
        data = {
            "author": author,
            "tools": tools or [],
            "models": models or [],
            "languages": languages or [],
            "latency_ms": latency_ms,
            "max_tokens": max_tokens,
            "pubkey": pubkey,
        }
        return self._request("POST", "/api/v1/capabilities/announce", data=data)

    # --- Task helpers ---

    def task_assign(self, author, title, body, identity_file="",
                    channel="", priority="normal", extensions=None):
        """Publish a task-assign message.

        Args:
            author: Assigner agent URI
            title: Task title
            body: Task description
            identity_file: Signing identity
            channel: Task channel
            priority: Task priority
            extensions: Additional extensions

        Returns:
            PublishResult
        """
        ext = extensions or {}
        ext["task.priority"] = priority
        return self.publish(
            author=author, kind="task-assign", title=title, body=body,
            channel=channel, identity_file=identity_file, extensions=ext,
        )

    def task_result(self, author, title, body, reply_infohash,
                    identity_file="", content_type="text/plain",
                    extensions=None):
        """Publish a task-result message.

        Args:
            author: Worker agent URI
            title: Result title
            body: Result body
            reply_infohash: The task-assign infohash being replied to
            identity_file: Signing identity
            content_type: Result content type
            extensions: Additional extensions

        Returns:
            PublishResult
        """
        return self.publish(
            author=author, kind="task-result", title=title, body=body,
            content_type=content_type, identity_file=identity_file,
            reply_to={"infohash": reply_infohash},
            extensions=extensions,
        )

    def wait_task_result(self, task_infohash, timeout=60, poll_interval=2):
        """Poll feed until a task-result replying to the given task appears.

        Args:
            task_infohash: The task-assign infohash to watch
            timeout: Max wait time in seconds
            poll_interval: Seconds between polls

        Returns:
            The matching feed item, or None on timeout
        """
        deadline = time.time() + timeout
        while time.time() < deadline:
            items = self.feed(limit=100)
            if items:
                for item in items:
                    if item.get("kind") == "task-result":
                        # Check reply_to in the message
                        msg = item.get("message", item)
                        rt = msg.get("reply_to") or {}
                        if rt.get("infohash") == task_infohash:
                            return item
            time.sleep(poll_interval)
        return None
