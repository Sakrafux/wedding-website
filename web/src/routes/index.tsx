import { createFileRoute } from "@tanstack/react-router";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export const Route = createFileRoute("/")({
  component: ScaffoldPage,
});

/**
 * A placeholder that exercises the design tokens, the fonts and the four shadcn
 * primitives, so that "the scaffold works" is something you can see rather than
 * infer from a green build. F1-F01 replaces it with the real login screen.
 */
function ScaffoldPage() {
  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-12 px-4 py-12 md:px-6">
      <header className="flex flex-col gap-4">
        <p className="text-small text-ink-muted">17. Juli 2027</p>
        <h1 className="text-display">Wir heiraten</h1>
        <p className="max-w-prose">
          Das Gerüst steht: Schriften, Farben und die ersten Bausteine sind da. Die echten Seiten kommen als Nächstes.
        </p>
      </header>

      <Card className="border-line bg-surface shadow-card rounded-xl">
        <CardHeader>
          <CardTitle className="text-h3">Dein Code</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="scaffold-code">Code von der Einladung</Label>
            <Input id="scaffold-code" placeholder="ABC-234" />
          </div>
          <Button className="w-full">Weiter</Button>
        </CardContent>
      </Card>
    </div>
  );
}
