import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppIllustrationPlaceholder from "@/components/app/AppIllustrationPlaceholder";

export default function GuildMarketplacePage() {
  return (
    <AppGuildLayout>
      <div>
        <div className="text-4xl font-bold text-white mb-4">Marketplace</div>
        <div className="text-lg font-light text-gray-300 mb-20">
          The plugin marketplace is a place where you can share your plugins and
          find ones made by others. Plugins on the marketplace are publicly
          available and can be installed by anyone.
        </div>
        <AppIllustrationPlaceholder
          svgPath="/illustrations/development.svg"
          title="Hey! You are early and we are still figuring things out here ..."
        />
      </div>
    </AppGuildLayout>
  );
}
