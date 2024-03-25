interface Props {
  title: string;
  description?: string;
}

export default function AppGuildPageHeader({ title, description }: Props) {
  return (
    <div>
      <h1 className="text-xl font-semibold md:text-2xl">{title}</h1>
      {description && (
        <div className="text-muted-foreground mt-2">{description}</div>
      )}
    </div>
  );
}
