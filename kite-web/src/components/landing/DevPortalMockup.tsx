const intents = [
  "Presence Intent",
  "Server Members Intent",
  "Message Content Intent",
];

// A simplified version of the Bot page in the Discord Developer Portal.
export default function DevPortalMockup() {
  return (
    <div className="rounded-lg border bg-[#313338] text-white p-6 space-y-6 not-prose">
      <div>
        <div className="text-xs font-bold uppercase text-gray-400 mb-2">
          Token
        </div>
        <div className="flex items-center gap-3">
          <div className="flex-auto rounded bg-[#1e1f22] px-3 py-2 text-gray-500 text-sm truncate">
            ••••••••••••••••••••••••••••••••
          </div>
          <div className="flex-none whitespace-nowrap rounded bg-[#5865F2] px-4 py-2 text-sm font-medium">
            Reset Token
          </div>
        </div>
      </div>
      <div>
        <div className="text-xs font-bold uppercase text-gray-400 mb-3">
          Privileged Gateway Intents
        </div>
        <div className="space-y-3">
          {intents.map((intent) => (
            <div key={intent} className="flex items-center justify-between">
              <div className="text-sm">{intent}</div>
              <div className="w-10 h-6 rounded-full bg-[#23a55a] flex items-center justify-end px-1">
                <div className="w-4 h-4 rounded-full bg-white" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
