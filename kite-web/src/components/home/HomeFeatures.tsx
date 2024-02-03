import clsx from "clsx";

const features = [
  {
    title: "Fearless Hosting",
    description:
      "Experience a new era of Discord Bot deployment with Kite's cutting-edge inbuilt cloud hosting and scaling feature. Say goodbye to complex server configurations and hello to instant deployment, allowing users to launch their Discord Bots seamlessly, ensuring reliability and performance at any scale.",
    illustration: "hosting",
  },
  {
    title: "Powerful Web Editor",
    description:
      "The Kite Web Code Editor is a powerful and user-friendly online platform designed to streamline the creation and deployment of Discord Bots directly from the web browser. Offering an intuitive interface and a rich set of features, this tool empowers users, from beginners to experienced developers, to effortlessly code, test, and deploy Discord Bots with unprecedented ease.",
    illustration: "code_editor",
  },
  {
    title: "Insightful Metrics",
    description:
      "Kite takes Discord Bot development to the next level with its comprehensive suite of insightful metrics. Empowering developers with valuable data-driven insights, this feature provides a deeper understanding of bot performance, user engagement, and overall effectiveness.",
    illustration: "analytics",
  },
  {
    title: "Share plugins with the Community",
    description:
      "Discover, explore, and enhance your Discord server experience with Kite's Marketplace—a vibrant ecosystem where users can access a diverse array of community-made Discord Bots. The Marketplace enriches your server's capabilities, providing a seamless platform for bot creators to share their innovations with the wider Discord community.",
    illustration: "marketplace",
  },
] as const;

type Feature = (typeof features)[number];

export default function HomeFeatures() {
  return (
    <div className="px-5 sm:px-10">
      <div className="py-32 space-y-20 lg:space-y-32 max-w-7xl mx-auto">
        {features.map((feature, i) => (
          <HomeFeature key={i} feature={feature} reverse={i % 2 === 1} />
        ))}
      </div>
    </div>
  );
}

function HomeFeature({
  feature,
  reverse,
}: {
  feature: Feature;
  reverse: boolean;
}) {
  return (
    <div
      className={clsx(
        "flex flex-col gap-20 lg:gap-32",
        reverse ? "lg:flex-row-reverse" : "lg:flex-row"
      )}
    >
      <div className="w-full">
        <h2 className="text-3xl font-semibold text-gray-100 mb-5">
          {feature.title}
        </h2>
        <div className="text-xl text-gray-300 leading-relaxed">
          {feature.description}
        </div>
      </div>
      <div className="w-full">
        <img src={`/illustrations/${feature.illustration}.svg`} alt="" />
      </div>
    </div>
  );
}
