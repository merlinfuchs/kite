// The parsing of secret references, like the service's in pkg/flow/secrets.go.
// Only placeholders reference secrets, and webhook.data.secrets.count, after
// a dot, isn't one.
const placeholderRe = /\{\{([\s\S]*?)\}\}/g;
const secretReferenceRe = /(?<![.\w])secrets\.([A-Za-z_][A-Za-z0-9_]*)/g;
const plainSecretRe = /^\s*secrets\.[A-Za-z_][A-Za-z0-9_]*\s*$/;
// Any use of the secrets, also like secrets?.NAME, which can read the ones
// another placeholder of the template references.
const secretMentionRe = /(?<![.\w])secrets\b/;

// The names of the secrets a template references, like API_KEY in
// {{secrets.API_KEY}}.
export function templateSecretNames(template: string) {
  const names: string[] = [];
  for (const [, placeholder] of template.matchAll(placeholderRe)) {
    for (const [, name] of placeholder.matchAll(secretReferenceRe)) {
      if (!names.includes(name)) names.push(name);
    }
  }
  return names;
}

// Whether every placeholder that references a secret is just the secret, like
// {{secrets.TOKEN}}, and not an expression that could use a part of it.
export function onlyPlainSecrets(template: string) {
  for (const [, placeholder] of template.matchAll(placeholderRe)) {
    if (secretMentionRe.test(placeholder) && !plainSecretRe.test(placeholder)) {
      return false;
    }
  }
  return true;
}
