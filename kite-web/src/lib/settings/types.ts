import { ComponentType } from "react";

export type MobileNavStyle = "docked" | "floating";
export type MobileNavTabs = "3-tabs" | "4-tabs" | "5-tabs";

export interface KiteSettings {
  mobileNavStyle: MobileNavStyle;
  mobileNavTabs: MobileNavTabs;
  showMobileLabels: boolean;
  [key: string]: any;
}

export const defaultKiteSettings: KiteSettings = {
  mobileNavStyle: "docked",
  mobileNavTabs: "4-tabs",
  showMobileLabels: true,
};

export interface KiteSettingSection {
  id: string;
  title: string;
  description: string;
  icon: ComponentType<{ className?: string }>;
  mobileOnly?: boolean;
  desktopOnly?: boolean;
  component: ComponentType;
}
