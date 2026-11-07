import { readFileSync } from "node:fs";
import { join } from "node:path";
import { generatedTokenFiles } from "../scripts/tokens.ts";

for (const [path, content] of Object.entries(generatedTokenFiles)) {
  it(`${path} matches src/tokens/source.ts`, () => {
    const current = readFileSync(join(process.cwd(), path), "utf8");
    expect(current, "Run `pnpm --filter paper-ui tokens`.").toBe(content);
  });
}
