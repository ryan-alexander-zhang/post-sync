const features = [
  "Markdown upload and parsing",
  "Telegram channel delivery",
  "Publish history and delivery status",
  "SQLite / PostgreSQL compatibility",
];

export default function HomePage() {
  return (
    <main className="mx-auto flex min-h-screen max-w-6xl flex-col gap-10 px-6 py-12">
      <section className="grid gap-6 rounded-[32px] border border-border bg-card/90 p-8 shadow-[0_20px_80px_rgba(18,91,80,0.12)] md:grid-cols-[1.4fr_0.8fr]">
        <div className="space-y-5">
          <p className="text-sm uppercase tracking-[0.3em] text-primary">
            Content distributor MVP
          </p>
          <h1 className="max-w-3xl text-5xl font-semibold leading-tight text-foreground">
            Publish one Markdown source to multiple external channels.
          </h1>
          <p className="max-w-2xl text-lg leading-8 text-foreground/75">
            This repository now contains the executable MVP blueprint and the
            implementation scaffold for backend APIs, delivery orchestration,
            and the Next.js admin console.
          </p>
        </div>
        <div className="rounded-[28px] bg-primary p-6 text-white">
          <p className="text-sm uppercase tracking-[0.25em] text-white/70">
            Current status
          </p>
          <div className="mt-6 space-y-4 text-sm leading-7">
            {features.map((feature) => (
              <div
                key={feature}
                className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3"
              >
                {feature}
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
