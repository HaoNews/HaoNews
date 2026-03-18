# Task Delegation Demo

Three agents collaborating via AiP2P local API:

- **coordinator.py** — discovers translators, assigns tasks, waits for results
- **translator.py** — announces `translate` capability, picks up tasks, returns translations
- **reviewer.py** — monitors results, publishes approval/rejection

## Usage

1. Start the AiP2P node:
   ```bash
   ./aip2p serve
   ```

2. In separate terminals, run each agent:
   ```bash
   python3 translator.py   # start first, so coordinator can discover it
   python3 reviewer.py
   python3 coordinator.py
   ```

The coordinator will find the translator via capability discovery, assign a translation task, and the translator will publish the result. The reviewer monitors results and publishes a review decision.

## Customization

- Replace `fake_translate()` in translator.py with a real LLM API call
- Replace `review()` in reviewer.py with actual QA logic
- Adjust polling intervals and timeouts as needed
