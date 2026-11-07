import { describe, expect, it } from "vitest";
// @ts-expect-error The plugin is plain ESM without declarations.
import { arbitraryValueProblem } from "../eslint-plugin.mjs";

describe("paper/no-arbitrary-values", () => {
  it("rejects values that bypass the tokens", () => {
    for (const candidate of ["p-[13px]", "bg-[#fff]", "hover:text-[1.1rem]", "[mask:url(x)]", "w-(--nope)", "w-[var(--nope)]", "!m-[2px]"]) {
      expect(arbitraryValueProblem(candidate), candidate).toBeTruthy();
    }
  });

  it("allows token utilities, token and runtime properties, and arbitrary selectors", () => {
    for (const candidate of [
      "p-4",
      "max-tablet:grid-cols-single",
      "w-(--anchor-width)",
      "origin-(--transform-origin)",
      "[background:var(--effect-table-scroll)]",
      "[&_a]:underline",
      "has-[>input:focus-visible]:outline-halo",
      "data-[orientation=vertical]:grid",
    ]) {
      expect(arbitraryValueProblem(candidate), candidate).toBeNull();
    }
  });
});
