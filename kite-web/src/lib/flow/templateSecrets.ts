// The parsing of secret and parameter references, like the service's in
// pkg/flow/secrets.go.
// Only placeholders reference secrets, and webhook.data.secrets.count, after
// a dot, isn't one.
const placeholderRe = /\{\{([\s\S]*?)\}\}/g;
const plainSecretRe = /^\s*secrets\.[A-Za-z_][A-Za-z0-9_]*\s*$/;
// Any use of the secrets, also like secrets?.NAME, which can read the ones
// another placeholder of the template references.
const secretMentionRe = /(?<![.\w])secrets\b/;

// The names of the secrets a template references, like API_KEY in
// {{secrets.API_KEY}}.
export function templateSecretNames(template: string) {
  return placeholderReferences(template, "secrets");
}

const referenceRes = new Map<string, RegExp>();

// placeholderReferences are the names the placeholders of a template use after
// a root like secrets or params, in the order they're first used.
export function placeholderReferences(template: string, root: string) {
  let referenceRe = referenceRes.get(root);
  if (!referenceRe) {
    referenceRe = new RegExp(
      `(?<![.\\w])${root}\\.([A-Za-z_][A-Za-z0-9_]*)`,
      "g"
    );
    referenceRes.set(root, referenceRe);
  }

  const names: string[] = [];
  for (const [, placeholder] of template.matchAll(placeholderRe)) {
    for (const [, name] of placeholder.matchAll(referenceRe)) {
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
