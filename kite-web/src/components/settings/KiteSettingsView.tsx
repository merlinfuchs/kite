import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useIsMobile } from "@/lib/hooks/use-mobile";
import { kiteSettingsSections } from "@/lib/settings/registry";
import { useEffect, useState } from "react";

interface Props {
  className?: string;
  forceShowMobile?: boolean;
}

export default function KiteSettingsView({
  className,
  forceShowMobile = false,
}: Props) {
  const isMobile = useIsMobile();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return null;
  }

  const visibleSections = kiteSettingsSections.filter((section) => {
    if (section.mobileOnly && !isMobile && !forceShowMobile) {
      return false;
    }
    if (section.desktopOnly && isMobile) {
      return false;
    }
    return true;
  });

  return (
    <div className={`space-y-6 ${className || ""}`}>
      {visibleSections.map((section) => {
        const Icon = section.icon;
        const SectionContent = section.component;

        return (
          <Card key={section.id}>
            <CardHeader>
              <div className="flex items-center gap-2">
                <Icon className="size-5 text-muted-foreground" />
                <CardTitle>{section.title}</CardTitle>
              </div>
              {section.description && (
                <CardDescription>{section.description}</CardDescription>
              )}
            </CardHeader>
            <CardContent>
              <SectionContent />
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
