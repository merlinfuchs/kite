import z from "zod";

export const clientEnvSchema = z.object({
  NEXT_PUBLIC_API_PUBLIC_BASE_URL: z.string().default("http://localhost:8080"),
  // Used for canonical URLs and social previews, which need absolute URLs.
  NEXT_PUBLIC_SITE_URL: z.string().default("https://kite.onl"),
  NEXT_PUBLIC_DOCS_LINK: z.string().default("http://localhost:4000"),
  NEXT_PUBLIC_GITHUB_LINK: z
    .string()
    .default("https://github.com/merlinfuchs/kite"),
  NEXT_PUBLIC_DISCORD_LINK: z.string().default("https://discord.gg"),
  NEXT_PUBLIC_CONTACT_EMAIL: z.string().default("contact@kite.onl"),
  // Analytics are only sent when set, since they need the /ops proxy in nginx.
  NEXT_PUBLIC_OPENPANEL_CLIENT_ID: z.string().optional(),
});

export default clientEnvSchema.parse({
  NEXT_PUBLIC_API_PUBLIC_BASE_URL: process.env.NEXT_PUBLIC_API_PUBLIC_BASE_URL,
  NEXT_PUBLIC_SITE_URL: process.env.NEXT_PUBLIC_SITE_URL,
  NEXT_PUBLIC_DOCS_LINK: process.env.NEXT_PUBLIC_DOCS_LINK,
  NEXT_PUBLIC_GITHUB_LINK: process.env.NEXT_PUBLIC_GITHUB_LINK,
  NEXT_PUBLIC_DISCORD_LINK: process.env.NEXT_PUBLIC_DISCORD_LINK,
  NEXT_PUBLIC_CONTACT_EMAIL: process.env.NEXT_PUBLIC_CONTACT_EMAIL,
  NEXT_PUBLIC_OPENPANEL_CLIENT_ID: process.env.NEXT_PUBLIC_OPENPANEL_CLIENT_ID,
});
