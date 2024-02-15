import clsx from "clsx";
import dynamic from "next/dynamic";
import { useState } from "react";

const CodeHiglight = dynamic(() => import("@/components/CodeHighlight"), {
  ssr: false,
});

const supportedLanguages = {
  typescript: {
    icon: "file_type_typescript",
    name: "TypeScript",
    exampleCode: `
import { call, event } from "@merlingg/kite-sdk";

event.on("DISCORD_MESSAGE_CREATE", (msg) => {
    if (msg.content === "!ping") {
        call("DISCORD_MESSAGE_CREATE", {
            channel_id: msg.channel_id,
            content: "Pong!",
        })
    }
})
    `.trim(),
  },
  go: {
    icon: "file_type_go",
    name: "Go",
    exampleCode: `
func main() {
	kite.Event(event.DiscordMessageCreate, func(req event.Event) error {
		msg := req.Data.(distype.MessageCreateEvent)

		if msg.Content == "!ping" {
			_, err := discord.MessageCreate(distype.MessageCreateRequest{
				ChannelID: msg.ChannelID,
				Content:   "Pong!",
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
}

    `.trim(),
  },
  rust: {
    icon: "file_type_rust",
    name: "Rust",
    exampleCode: null,
  },
  python: {
    icon: "file_type_python",
    name: "Python",
    exampleCode: null,
  },
} as const;

export type SupportedLanguageKey = keyof typeof supportedLanguages;

export default function HomeCodeShowcase() {
  const [selectedLanguage, setSelectedLanguage] =
    useState<SupportedLanguageKey>("typescript");
  const language = supportedLanguages[selectedLanguage];

  return (
    <div className="flex flex-col gap-10">
      <div className="w-[700px] h-[600px] bg-dark-2 mx-auto rounded-xl p-3 overflow-hidden">
        {language.exampleCode ? (
          <div className="flex flex-col space-y-3 h-full">
            <div className="overflow-y-auto px-4 py-3 bg-dark-3 rounded-lg flex-auto">
              <CodeHiglight
                code={language.exampleCode!}
                language={selectedLanguage}
              ></CodeHiglight>
            </div>
            <div className="flex-none h-32 bg-dark-1 px-4 py-3 rounded-lg font-mono space-y-3">
              <div className="flex space-x-2 items-center text-gray-400">
                <div>{"> kite plugin init --type " + selectedLanguage}</div>
              </div>
              <div className="flex space-x-2 items-center text-gray-300">
                <div>
                  {"> kite plugin deploy --guild_id 615613572164091914"}
                </div>
                <div className="h-5 w-2 bg-gray-300 animate-pulse"></div>
              </div>
            </div>
          </div>
        ) : (
          <div className="px-20 overflow-hidden">
            <img src="/illustrations/wip.svg" alt="" className="py-10" />
            <div className="text-xl text-gray-300 font-light text-center">
              <span className="font-medium text-gray-100">{language.name}</span>{" "}
              support is coming soon!
            </div>
          </div>
        )}
      </div>
      <div className="flex justify-center gap-10 items-center">
        {Object.entries(supportedLanguages).map(([id, lang]) => (
          <SupportedLanguage
            key={id}
            name={lang.name}
            icon={lang.icon}
            selected={id === selectedLanguage}
            onClick={() => setSelectedLanguage(id as SupportedLanguageKey)}
          />
        ))}
      </div>
    </div>
  );
}

function SupportedLanguage({
  icon,
  name,
  selected,
  onClick,
}: {
  icon: string;
  name: string;
  selected: boolean;
  onClick?: () => void;
}) {
  return (
    <div
      className={clsx(
        "rounded-xl flex items-center justify-center h-20 w-20 cursor-pointer transition-all",
        selected ? "bg-dark-1 scale-110" : "bg-dark-2"
      )}
      aria-label={name}
      onClick={onClick}
    >
      <img src={`/file-icons/${icon}.svg`} alt="" className="h-16 w-16" />
    </div>
  );
}
