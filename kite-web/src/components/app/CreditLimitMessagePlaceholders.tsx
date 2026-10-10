const placeholders = [
  ["{{limit.credits}}", "the limit"],
  ["{{limit.used}}", "credits used"],
  ["{{limit.period}}", "today or this month"],
  ["{{limit.scope}}", "server or user"],
  ["{{limit.resets}}", "when it resets, like in 5 hours"],
  ["{{limit.resets_at}}", "the date and time it resets"],
  ["{{user.mention}}", "the user"],
  ["{{server.name}}", "the server"],
];

// Lists the placeholders a credit limit message can use.
export default function CreditLimitMessagePlaceholders() {
  return (
    <div className="text-xs text-muted-foreground space-y-1">
      <div>
        Shown privately to users who run a command or click a button after the
        limit is reached. Placeholders:
      </div>
      <div className="flex flex-wrap gap-x-3 gap-y-1">
        {placeholders.map(([placeholder, description]) => (
          <span key={placeholder} title={description}>
            <code className="text-foreground">{placeholder}</code>
          </span>
        ))}
      </div>
    </div>
  );
}
