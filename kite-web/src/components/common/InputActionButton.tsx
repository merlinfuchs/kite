import { LucideIcon } from "lucide-react";
import { ButtonHTMLAttributes, forwardRef } from "react";

// Small icon button in the label row of a text input, like the placeholder
// explorer and the emoji picker. Popovers use it as their trigger.
const InputActionButton = forwardRef<
  HTMLButtonElement,
  ButtonHTMLAttributes<HTMLButtonElement> & { label: string; icon: LucideIcon }
>(({ label, icon: Icon, ...props }, ref) => (
  <button
    ref={ref}
    type="button"
    title={label}
    aria-label={label}
    className="h-6 w-6 flex items-center justify-center rounded-md text-muted-foreground hover:text-foreground hover:bg-muted"
    {...props}
  >
    <Icon className="h-4 w-4" />
  </button>
));
InputActionButton.displayName = "InputActionButton";

export default InputActionButton;
