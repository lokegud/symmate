# tiers

Portable visitor access tiers for a digital-twin style assistant.

- **Tiers** (cumulative): 1 blocked, 2 public (default for anonymous), 3 schedule, 4 all.
- **Verification:** `Verifier` checks a visitor's facts (birthday, phone, email) against
  stored SHA-256 hashes. Every fact must match. Nothing here calls a model.
- **Grants:** `GrantsFor(level)` gives the data-class permissions for a tier.
- **Streaming filter:** `FilterStream` runs each complete sentence of a streamed reply
  through a `Redactor` (plug in your PII firewall) before it is sent.
- **Disclosure:** `PromptBlock(level, disclosure)` builds the prompt block that
  shares owner data, only at tier 3 and above.
- **Meeting requests:** `MeetingStore` keeps pending requests in a JSON file.

The package has no dependencies outside the standard library. Wire it to HTTP,
sessions and a model in the host project.
