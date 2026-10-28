import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import env from "../env/server";
import {
  installRequest,
  installSelection,
  itemLabel,
  updateSelection,
} from "../flow/templateInstall";
import {
  parameterDefaults,
  Template,
  templateFromBundle,
  templatePath,
} from "../flow/templates";
import {
  enabledIntegrationIds,
  missingFlowIntegrations,
} from "../integrations/missing";
import { matchesSearch } from "../search";
import { MessageData } from "../types/message.gen";
import {
  AppIntegration,
  Bundle,
  BundleInstall,
  BundleInstallUpdateRequest,
  MessageInstanceCreateRequest,
  UserBundle,
} from "../types/wire.gen";
import { APICaller, appId, appPath, json, maxSearchResults } from "./util";
import { parseMessage } from "./write";

// The tools for the templates of the marketplace. Adding and updating them is
// like doing it on the template's page, publishing them needs a login.
export function registerTemplateTools(
  server: McpServer,
  { get, list, send }: APICaller,
  canWrite: boolean
) {
  const read = { readOnlyHint: true, openWorldHint: false };
  const templateId = z
    .string()
    .describe("The ID of the template, from search_templates.");

  const getTemplate = async (id: string) => {
    const bundle = await get<Bundle>(`/v1/bundles/${encodeURIComponent(id)}`);
    const template = templateFromBundle(bundle);
    if (!template) throw new Error(`Template ${id} can't be added.`);
    return template;
  };
  const getInstall = async (appId: string, templateId: string) => {
    const installs = await list<BundleInstall>(
      appPath(appId, "bundle-installs")
    );
    return installs.find((i) => i.bundle_id === templateId);
  };

  server.registerTool(
    "search_templates",
    {
      title: "Search templates",
      description:
        "Searches the templates of the marketplace, made by Kite or by users, and the user's own ones. A template adds commands, event listeners and messages to an app, which is quicker than building them. Without words it lists all of them.",
      inputSchema: {
        query: z
          .string()
          .optional()
          .describe('Words to search for, like "tickets".'),
      },
      annotations: read,
    },
    async ({ query }) => {
      const [bundles, own] = await Promise.all([
        list<Bundle>("/v1/bundles"),
        list<UserBundle>("/v1/users/@me/bundles"),
      ]);
      const listed = bundles.flatMap((bundle) => {
        const template = templateFromBundle(bundle);
        return template ? [templateInfo(template)] : [];
      });
      const ids = new Set(listed.map((t) => t.id));
      const results = [
        ...listed,
        ...own
          .filter((b) => !ids.has(b.id))
          .map((b) => ({
            id: b.id,
            name: b.name,
            description: b.description,
            made_by: "you",
            version: b.version,
            installs: b.install_count,
            url: undefined,
          })),
      ].filter((t) =>
        matchesSearch(query ?? "", [t.id, t.name, t.description])
      );
      return json(results.slice(0, maxSearchResults));
    }
  );

  server.registerTool(
    "get_template",
    {
      title: "Get template",
      description:
        "Gets a template: what it adds, the parameters to fill in when adding it, like a channel, and the integrations it needs. With app_id, also what the app has of it already and what updating it would change.",
      inputSchema: {
        template_id: templateId,
        app_id: appId.optional(),
      },
      annotations: read,
    },
    async ({ template_id, app_id }) => {
      const template = await getTemplate(template_id);
      const res: Record<string, unknown> = {
        ...templateInfo(template),
        about: template.about,
        setup: template.setup,
        parameters: template.parameters.map((p) => ({
          key: p.key,
          label: p.label,
          description: p.description,
          type: p.type,
          required: p.required,
          default: p.default,
        })),
        commands: template.commands.map((c) => ({
          key: c.key,
          name: c.name,
          description: c.description,
          parameter_keys: c.parameterKeys,
        })),
        event_listeners: template.eventListeners.map((l) => ({
          key: l.key,
          event: itemLabel(template, l.key),
          description: l.description,
          parameter_keys: l.parameterKeys,
        })),
        messages: template.messages.map((m) => ({
          key: m.key,
          name: m.name,
          description: m.description ?? undefined,
          parameter_keys: m.parameterKeys,
          posted_in: m.sendTo,
          written_by_user: m.customize || undefined,
        })),
        variables: template.variables.map((v) => ({
          key: v.key,
          name: v.name,
        })),
      };

      if (app_id) {
        const [install, states] = await Promise.all([
          getInstall(app_id, template.id),
          list<AppIntegration>(appPath(app_id, "integrations")),
        ]);
        res.missing_integrations = missingFlowIntegrations(
          itemFlows(template),
          enabledIntegrationIds(states)
        ).map((i) => ({ id: i.id, name: i.name }));

        if (install) {
          const update = updateSelection(template, install, new Set());
          res.installed = {
            version: install.bundle_version,
            parameters: install.parameters,
            items: install.items.map((i) => ({
              key: i.key,
              edited: i.edited || undefined,
              deleted: i.deleted || undefined,
            })),
            update: update.outdated
              ? {
                  version_notes: template.versionNotes.filter(
                    (n) => n.version > install.bundle_version
                  ),
                  edited_items: update.edited.map((i) => i.key),
                  new_parameters: update.missingParameters.map((p) => p.key),
                }
              : undefined,
          };
        }
      }
      return json(res);
    }
  );

  if (!canWrite) return;

  server.registerTool(
    "install_template",
    {
      title: "Add template",
      description:
        "Adds a template to an app: its commands, event listeners, messages and stored variables, with the parameters filled in. Messages the template posts in a channel are posted right away. Templates not made by Kite are made by other users, so tell the user who made it and ask before adding one. Get the template with get_template and app_id first, and ask the user for the parameters you can't pick from list_channels and list_roles. When the app has a command with the same name, the template's command gets a number, like /giveaway-2, unless the app's command has nothing but its entry block, which gets the number instead.",
      inputSchema: {
        app_id: appId,
        template_id: templateId,
        items: z
          .array(z.string())
          .optional()
          .describe(
            "The keys of the commands and event listeners to add. Defaults to all of them. Messages and stored variables are always added, the others use them."
          ),
        parameters: z
          .record(z.string(), z.string())
          .default({})
          .describe(
            "Values of the parameters by their key. Channels and roles are their IDs."
          ),
        server_id: z
          .string()
          .optional()
          .describe(
            "The ID of the server, from list_servers, the channels of the parameters are in. Needed when a message is posted."
          ),
        messages: z
          .record(z.string(), z.record(z.string(), z.unknown()))
          .optional()
          .describe(
            "For messages with written_by_user, the message like Discord's API takes it, by the message's key. Without it, it's the template's text."
          ),
      },
      annotations: {
        readOnlyHint: false,
        destructiveHint: false,
        openWorldHint: false,
      },
    },
    async ({ app_id, template_id, items, server_id, messages, ...args }) => {
      const [template, install] = await Promise.all([
        getTemplate(template_id),
        getInstall(app_id, template_id),
      ]);
      const parameters = {
        ...parameterDefaults(template.parameters),
        ...args.parameters,
      };
      if (install && install.bundle_version < template.version) {
        throw new Error(
          `The app has version ${install.bundle_version} of the template, update it to version ${template.version} with update_template_install first.`
        );
      }

      const optional = [...template.commands, ...template.eventListeners].map(
        (i) => i.key
      );
      const unknown = (items ?? []).filter((key) => !optional.includes(key));
      if (unknown.length) {
        throw new Error(
          `The template has no commands or event listeners ${unknown.join(
            ", "
          )}.`
        );
      }
      const skipped = new Set(
        items ? optional.filter((key) => !items.includes(key)) : []
      );
      const selection = installSelection(template, install, skipped);
      if (
        !selection.commands.length &&
        !selection.eventListeners.length &&
        !selection.messages.length &&
        !selection.variables.length
      ) {
        throw new Error("The app has all of it already.");
      }

      const missing = selection.parameters.filter(
        (p) => p.required && !parameters[p.key]
      );
      if (missing.length) {
        throw new Error(
          `Missing values for the parameters ${missing
            .map((p) => p.key)
            .join(", ")}.`
        );
      }
      const posted = selection.messages.filter(
        (m) => m.sendTo && parameters[m.sendTo]
      );
      if (posted.length && !server_id) {
        throw new Error(
          "server_id is needed to post the messages of the template."
        );
      }

      const customMessages: Record<string, MessageData> = {};
      for (const [key, data] of Object.entries(messages ?? {})) {
        if (!selection.customizable.some((m) => m.key === key)) {
          throw new Error(
            `The template adds no message ${key} that's written by the user.`
          );
        }
        customMessages[key] = parseMessage(data);
      }

      const request = installRequest(
        template,
        selection,
        parameters,
        customMessages
      );
      const res = await send<BundleInstall>(
        "POST",
        appPath(app_id, "bundle-installs"),
        request
      );

      // Messages are posted after they were created, like on the template's
      // page. One that can't be posted can be posted from Kite.
      const unposted: { message: string; error: string; post_url: string }[] =
        [];
      for (const message of posted) {
        const item = res.items.find(
          (i) => i.type === "message" && i.key === message.key
        );
        if (!item) continue;
        try {
          await send(
            "POST",
            appPath(app_id, "messages", item.id, "instances"),
            {
              discord_guild_id: server_id!,
              discord_channel_id: parameters[message.sendTo!],
            } satisfies MessageInstanceCreateRequest
          );
        } catch (err) {
          unposted.push({
            message: message.key,
            error: err instanceof Error ? err.message : String(err),
            post_url: `${env.NEXT_PUBLIC_SITE_URL}/apps/${app_id}/messages/${item.id}`,
          });
        }
      }

      const states = await list<AppIntegration>(
        appPath(app_id, "integrations")
      );
      const missingIntegrations = missingFlowIntegrations(
        itemFlows(selection),
        enabledIntegrationIds(states)
      );

      const added = new Set(request.items);
      return json({
        install_id: res.id,
        version: res.bundle_version,
        added: res.items
          .filter((i) => added.has(i.key))
          .map((i) => ({ key: i.key, type: i.type, id: i.id })),
        unposted: unposted.length ? unposted : undefined,
        // Their blocks fail until the user enables them in Kite.
        missing_integrations: missingIntegrations.length
          ? missingIntegrations.map((i) => ({ id: i.id, name: i.name }))
          : undefined,
      });
    }
  );

  server.registerTool(
    "update_template_install",
    {
      title: "Update template",
      description:
        "Updates what an app has of a template to the template's newest version. Items the user didn't edit are replaced, edited ones only if they're in overwrite, which loses the user's changes, so ask first. get_template with app_id shows what an update changes.",
      inputSchema: {
        app_id: appId,
        template_id: templateId,
        overwrite: z
          .array(z.string())
          .default([])
          .describe("Keys of edited items to replace with the new version."),
        parameters: z
          .record(z.string(), z.string())
          .default({})
          .describe(
            "Values of the parameters the new version needs, by their key."
          ),
      },
      annotations: {
        readOnlyHint: false,
        destructiveHint: true,
        openWorldHint: false,
      },
    },
    async ({ app_id, template_id, overwrite, ...args }) => {
      const [template, install] = await Promise.all([
        getTemplate(template_id),
        getInstall(app_id, template_id),
      ]);
      const parameters = {
        ...parameterDefaults(template.parameters),
        ...args.parameters,
      };
      if (!install) throw new Error("The app doesn't have this template.");

      const update = updateSelection(template, install, new Set(overwrite));
      if (!update.outdated) {
        throw new Error(
          `The app has the newest version ${install.bundle_version} already.`
        );
      }
      const notEdited = overwrite.filter(
        (key) => !update.edited.some((i) => i.key === key)
      );
      if (notEdited.length) {
        throw new Error(
          `${notEdited.join(", ")} aren't edited items the update keeps.`
        );
      }
      const missing = update.missingParameters.filter(
        (p) => p.required && !parameters[p.key]
      );
      if (missing.length) {
        throw new Error(
          `Missing values for the parameters ${missing
            .map((p) => p.key)
            .join(", ")}.`
        );
      }

      const res = await send<BundleInstall>(
        "POST",
        appPath(app_id, "bundle-installs", install.id, "update"),
        {
          bundle_version: template.version,
          parameters: Object.fromEntries(
            update.missingParameters.flatMap((p) =>
              parameters[p.key] ? [[p.key, parameters[p.key]]] : []
            )
          ),
          overwrite,
        } satisfies BundleInstallUpdateRequest
      );
      return json({
        version: res.bundle_version,
        replaced: update.updatable
          .filter((i) => !i.edited || overwrite.includes(i.key))
          .map((i) => i.key),
        kept: update.edited
          .filter((i) => !overwrite.includes(i.key))
          .map((i) => i.key),
      });
    }
  );
}

// What search_templates and get_template show of a template.
function templateInfo(template: Template) {
  return {
    id: template.id,
    name: template.name,
    description: template.description,
    made_by:
      template.author.kind === "kite"
        ? "Kite"
        : template.author.kind === "you"
        ? "you"
        : `${template.author.name} (@${template.author.username})`,
    version: template.version,
    installs: template.installs,
    url: template.listed
      ? `${env.NEXT_PUBLIC_SITE_URL}${templatePath(template.id)}`
      : undefined,
  };
}

// The flows of commands, event listeners and messages of a template.
function itemFlows(
  items: Pick<Template, "commands" | "eventListeners" | "messages">
) {
  return [
    ...[...items.commands, ...items.eventListeners].map((i) => i.flowSource),
    ...items.messages.flatMap((m) => Object.values(m.flowSources)),
  ];
}
