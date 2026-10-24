import { loadTemplate, Template, TemplateFile } from "./templates";
import askAi from "../templates/ask-ai.json";
import counting from "../templates/counting.json";
import erlc from "../templates/erlc.json";
import moderation from "../templates/moderation.json";
import modmail from "../templates/modmail.json";
import serverInfo from "../templates/server-info.json";
import starboard from "../templates/starboard.json";
import suggestions from "../templates/suggestions.json";
import verification from "../templates/verification.json";
import welcomer from "../templates/welcomer.json";

// The templates in the order of the Templates page. Each file is described by
// templateFileSchema in templates.ts, see "Adding a template" in AGENTS.md.
export const templateFiles: unknown[] = [
  moderation,
  askAi,
  welcomer,
  verification,
  suggestions,
  starboard,
  serverInfo,
  modmail,
  counting,
  erlc,
];

// The templates of the site, which its public pages are built from. The app
// loads the templates from the service.
const templates = (templateFiles as TemplateFile[]).map(loadTemplate);

export function getTemplates(): Template[] {
  return templates;
}

// templatePaths are the public pages of the templates.
export function templatePaths() {
  return templates.map((t) => `/templates/${t.id}`);
}

export function getTemplate(id: string) {
  return templates.find((t) => t.id === id);
}
