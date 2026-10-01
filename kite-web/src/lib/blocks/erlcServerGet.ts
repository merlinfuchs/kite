import { z } from "zod";
import { BlockDefinition, BlockField } from "./types";

// Parts of the server ER:LC only returns when they're asked for.
const parts: [string, string, string, string][] = [
  ["include_players", "Players", "Players", "the players in the server"],
  ["include_staff", "Staff", "Staff", "the server's admins, mods and helpers"],
  ["include_queue", "Queue", "Queue", "the players waiting to join"],
  [
    "include_join_logs",
    "JoinLogs",
    "Join Logs",
    "who joined and left recently",
  ],
  ["include_kill_logs", "KillLogs", "Kill Logs", "who killed whom recently"],
  [
    "include_command_logs",
    "CommandLogs",
    "Command Logs",
    "the commands run recently",
  ],
  [
    "include_mod_calls",
    "ModCalls",
    "Mod Calls",
    "recent calls for a moderator",
  ],
  [
    "include_emergency_calls",
    "EmergencyCalls",
    "Emergency Calls",
    "recent emergency calls",
  ],
  [
    "include_vehicles",
    "Vehicles",
    "Vehicles",
    "the vehicles spawned in the server",
  ],
];

const partFields: BlockField[] = parts.map(([name, target, label, what]) => ({
  name,
  in: "query",
  target,
  type: "boolean",
  label,
  description: `Whether to include ${what}.`,
}));

const player = z
  .string()
  .describe('Roblox name and ID of a player, as "name:id"');

export const erlcServerGet: BlockDefinition = {
  type: "action_erlc_server_get",
  title: "Get ER:LC server",
  description: "Get the status of your ER:LC private server",
  icon: "siren",
  category: "ER:LC",
  credits: 1,
  strict_settings: true,
  run: {
    kind: "request",
    integration: "erlc",
    operation: "get_server",
    method: "GET",
    path: "/v2/server",
  },
  fields: partFields,
  result: {
    schema: z
      .object({
        Name: z.string().describe("Name of the server"),
        OwnerId: z.number().describe("Roblox ID of the owner"),
        CoOwnerIds: z.array(z.number()).describe("Roblox IDs of the co-owners"),
        CurrentPlayers: z.number().describe("Number of players in the server"),
        MaxPlayers: z.number().describe("Number of players that fit"),
        JoinKey: z.string().describe("Code to join the server with"),
        AccVerifiedReq: z
          .string()
          .describe("Verification players need: Disabled, Email or Phone/ID"),
        TeamBalance: z.boolean().describe("Whether teams are balanced"),
        Players: z
          .array(
            z.object({
              Player: player,
              Team: z.string().describe("Team, like Police or Civilian"),
              Callsign: z
                .string()
                .nullable()
                .describe("Callsign, only on non-civilian teams"),
              Permission: z
                .string()
                .describe(
                  "Normal, Server Moderator, Server Administrator or Server Owner"
                ),
              WantedStars: z.number().describe("Wanted level"),
              Location: z
                .object({
                  PostalCode: z.string(),
                  StreetName: z.string(),
                  BuildingNumber: z.string(),
                })
                .describe("Where the player is"),
            })
          )
          .optional()
          .describe("Players in the server, if included"),
        Staff: z
          .object({
            Admins: z.record(z.string()),
            Mods: z.record(z.string()),
            Helpers: z.record(z.string()),
          })
          .optional()
          .describe(
            "Staff of the server, if included, each a map of Roblox ID to name"
          ),
        Queue: z
          .array(z.number())
          .optional()
          .describe("Roblox IDs of the players waiting to join, if included"),
        JoinLogs: z
          .array(
            z.object({
              Player: player,
              Join: z.boolean().describe("True for a join, false for a leave"),
              Timestamp: z.number().describe("Unix timestamp"),
            })
          )
          .optional()
          .describe("Recent joins and leaves, if included"),
        KillLogs: z
          .array(
            z.object({
              Killer: player,
              Killed: player,
              Timestamp: z.number().describe("Unix timestamp"),
            })
          )
          .optional()
          .describe("Recent kills, if included"),
        CommandLogs: z
          .array(
            z.object({
              Player: player,
              Command: z.string().describe("The command, like :h hello"),
              Timestamp: z.number().describe("Unix timestamp"),
            })
          )
          .optional()
          .describe("Recent commands, if included"),
        ModCalls: z
          .array(
            z.object({
              Caller: player,
              Moderator: player.optional(),
              Timestamp: z.number().describe("Unix timestamp"),
            })
          )
          .optional()
          .describe("Recent calls for a moderator, if included"),
        EmergencyCalls: z
          .array(
            z.object({
              Team: z.string().describe("Team called, like Police"),
              Caller: z.number().describe("Roblox ID of the caller"),
              CallNumber: z.number(),
              Description: z.string(),
              PositionDescriptor: z.string().describe("Where the caller is"),
              StartedAt: z.number().describe("Unix timestamp"),
            })
          )
          .optional()
          .describe("Recent emergency calls, if included"),
        Vehicles: z
          .array(
            z.object({
              Name: z.string().describe("Model, like Redline Fire Engine"),
              Owner: z.string().describe("Roblox name of the owner"),
              Plate: z.string(),
              Texture: z.string().nullable().describe("Livery"),
              ColorName: z.string(),
              ColorHex: z.string(),
            })
          )
          .optional()
          .describe("Spawned vehicles, if included"),
      })
      .describe("The server"),
  },
};
