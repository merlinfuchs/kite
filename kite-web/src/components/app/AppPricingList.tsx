import { CheckIcon, InfinityIcon } from "lucide-react";
import { Button } from "../ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Badge } from "../ui/badge";
import { useAppSubscriptions } from "@/lib/hooks/api";
import { ReactNode } from "react";

interface PricingProps {
  title: string;
  popular: boolean;
  current: boolean;
  price: number;
  description: string;
  benefitList: string[];
}

const pricingList: PricingProps[] = [
  {
    title: "Basic",
    popular: false,
    current: true,
    price: 0,
    description: "Get started with Kite for free. No credit card required.",
    benefitList: [
      "1 Collaborator",
      "10,000 credits / month",
      "Up to 25 servers",
      "Community support",
    ],
  },
  {
    title: "Premium",
    popular: true,
    current: false,
    price: 5.99,
    description: "Get more out of Kite with the premium plan for your app.",
    benefitList: [
      "5 Collaborators",
      "100,000 credits / month",
      "Up to 1,000 servers",
      "Priority support",
    ],
  },
  {
    title: "Ultimate",
    popular: false,
    current: false,
    price: 14.99,
    description:
      "Get the most out of Kite with the ultimate plan for your app.",
    benefitList: [
      "25 Collaborators",
      "1,000,000 credits / month",
      "Up to 2,500 servers",
      "Priority support",
    ],
  },
];

export default function AppPricingList() {
  const subscriptions = useAppSubscriptions();

  const activeSubscription = subscriptions?.find(
    (subscription) => subscription!.status !== "expired"
  );

  console.log(activeSubscription);

  return (
    <div className="grid lg:grid-cols-2 xl:grid-cols-3 gap-8 xl:mx-16">
      {pricingList.map((pricing: PricingProps) => (
        <Card
          key={pricing.title}
          className={
            pricing.popular
              ? "drop-shadow-xl shadow-black/10 dark:shadow-white/10"
              : "xl:my-8 "
          }
        >
          <CardHeader>
            <CardTitle className="flex item-center justify-between">
              {pricing.title}
              {pricing.popular ? (
                <Badge variant="secondary" className="text-sm text-primary">
                  Best Value
                </Badge>
              ) : null}
            </CardTitle>
            <div>
              <span className="text-3xl font-bold">${pricing.price}</span>
              <span className="text-muted-foreground"> /month</span>
            </div>

            <CardDescription>{pricing.description}</CardDescription>
          </CardHeader>

          <CardContent>
            <Button
              className="w-full"
              disabled={pricing.current}
              variant={pricing.popular ? "default" : "outline"}
            >
              {pricing.current ? "Current Plan" : "Get Started"}
            </Button>
          </CardContent>

          <hr className="w-4/5 m-auto mb-4" />

          <CardFooter className="flex">
            <div className="space-y-4">
              {pricing.benefitList.map((benefit) => (
                <span key={benefit} className="flex">
                  <CheckIcon className="text-green-500" />{" "}
                  <h3 className="ml-2">{benefit}</h3>
                </span>
              ))}
            </div>
          </CardFooter>
        </Card>
      ))}
    </div>
  );
}
