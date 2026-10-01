import AppearanceSection from "@/components/settings/sections/AppearanceSection";
import MobileNavigationSection from "@/components/settings/sections/MobileNavigationSection";
import { KiteSettingSection } from "@/lib/settings/types";
import { PaletteIcon, SmartphoneIcon } from "lucide-react";

export const kiteSettingsSections: KiteSettingSection[] = [
  {
    id: "mobile-navigation",
    title: "Mobile Navigation",
    description:
      "Customize the layout, tabs, and appearance of the mobile bottom navigation bar.",
    icon: SmartphoneIcon,
    mobileOnly: true,
    component: MobileNavigationSection,
  },
  {
    id: "appearance",
    title: "Appearance",
    description: "Personalize the interface theme and styling for Kite.",
    icon: PaletteIcon,
    component: AppearanceSection,
  },
];
