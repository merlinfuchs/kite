// The domain of an OAuth client, from its ID, the URL of its metadata
// document. Only the owner of the domain can host it, so it's who the client
// is, while its name can be anything.
export function oauthClientDomain(clientId: string) {
  try {
    return new URL(clientId).host;
  } catch {
    return clientId;
  }
}

// Domains of the clients Kite knows, like Claude and ChatGPT, which the
// consent page marks as verified.
const verifiedClientDomains = [
  "claude.ai",
  "chatgpt.com",
  "openai.com",
  "vscode.dev",
];

export function isVerifiedClientDomain(domain: string) {
  return verifiedClientDomains.includes(domain);
}
