import IllustrationPlaceholder from "./IllustrationPlaceholder";

export default function LoginPrompt() {
  return (
    <div className="my-32 lg:my-48 px-5">
      <IllustrationPlaceholder
        className="mb-10"
        svgPath="/illustrations/signin.svg"
        title="Login with your Discord account to access Kite and get started!"
      />
      <div className="text-center">
        <a
          href="/api/v1/auth/redirect"
          className="px-5 py-3 bg-primary rounded text-white text-lg"
        >
          Login with Discord
        </a>
      </div>
    </div>
  );
}
