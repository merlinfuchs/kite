import {
  loadTemplate,
  Template,
  TemplateFile,
  templatePath,
} from "./templates";
import afk from "../templates/afk.json";
import aiAutomod from "../templates/ai-automod.json";
import askAi from "../templates/ask-ai.json";
import counting from "../templates/counting.json";
import erlc from "../templates/erlc.json";
import giveaway from "../templates/giveaway.json";
import honeypot from "../templates/honeypot.json";
import moderation from "../templates/moderation.json";
import modmail from "../templates/modmail.json";
import reminders from "../templates/reminders.json";
import serverInfo from "../templates/server-info.json";
import starboard from "../templates/starboard.json";
import stickyMessage from "../templates/sticky-message.json";
import suggestions from "../templates/suggestions.json";
import verification from "../templates/verification.json";
import welcomer from "../templates/welcomer.json";

// The templates in the order of the Marketplace page. Each file is described by
// templateFileSchema in templates.ts, see "Adding a template" in AGENTS.md.
export const templateFiles: unknown[] = [
  moderation,
  aiAutomod,
  askAi,
  welcomer,
  verification,
  suggestions,
  starboard,
  serverInfo,
  modmail,
  counting,
  erlc,
  afk,
  giveaway,
  stickyMessage,
  reminders,
  honeypot,
];

// The templates of the site, which its public pages are built from. The app
// loads the templates from the service.
const templates = (templateFiles as TemplateFile[]).map((file) =>
  loadTemplate(file, { kind: "kite" }, true)
);

export function getTemplates(): Template[] {
  return templates;
}

export function templatePaths() {
  return templates.map((t) => templatePath(t.id));
}

export function getTemplate(id: string) {
  return templates.find((t) => t.id === id);
}
