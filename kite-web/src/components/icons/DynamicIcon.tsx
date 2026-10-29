import dynamic from "next/dynamic";
import { LucideProps } from "lucide-react";
import dynamicIconImports from "lucide-react/dynamicIconImports";
import { ComponentType } from "react";

type IconName = keyof typeof dynamicIconImports;

interface IconProps extends LucideProps {
  name: IconName;
}

// Created once per icon, a new component on every render would remount it.
// Not server-rendered, the server may have loaded an icon the client hasn't
// yet, which breaks hydration.
const icons = new Map<IconName, ComponentType<LucideProps>>();

function getIcon(name: IconName) {
  let icon = icons.get(name);
  if (!icon) {
    icon = dynamic(dynamicIconImports[name], { ssr: false });
    icons.set(name, icon);
  }
  return icon;
}

export default function DynamicIcon({ name, ...props }: IconProps) {
  const LucideIcon = getIcon(name);

  // getIcon returns the same component for a name on every render.
  // eslint-disable-next-line react-hooks/static-components
  return <LucideIcon {...props} />;
}
