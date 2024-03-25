import AppGuildLayout from "@/components/app/AppGuildLayout";
import AppGuildPageEmpty from "@/components/app/AppGuildPageEmpty";
import AppGuildPageHeader from "@/components/app/AppGuildPageHeader";
import AppIllustrationPlaceholder from "@/components/app/AppIllustrationPlaceholder";

export default function GuildMarketplacePage() {
  return (
    <AppGuildLayout>
      <AppGuildPageHeader
        title="Marketplace"
        description="The plugin marketplace is a place where you can share your plugins and
          find ones made by others. Plugins on the marketplace are publicly
          available and can be installed by anyone."
      />
      <AppGuildPageEmpty
        title="Under Construction"
        description="The plugin market place is still being worked on and will be ready soon!"
      />
    </AppGuildLayout>
  );
}
