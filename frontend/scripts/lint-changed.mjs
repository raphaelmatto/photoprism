import { execFileSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { ESLint } from "eslint";

const frontendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const baseRef = process.argv[2];

if (!baseRef || process.argv.length !== 3) {
  console.error("Usage: node scripts/lint-changed.mjs <base-ref>");
  process.exit(2);
}

// Resolve the ref before using it so a missing baseline fails the check.
const baseSha = execFileSync("git", ["rev-parse", "--verify", "--end-of-options", `${baseRef}^{commit}`], {
  cwd: frontendDir,
  encoding: "utf8",
}).trim();

const files = execFileSync("git", ["diff", "--relative", "--name-only", "--diff-filter=ACMR", "-z", baseSha, "--"], {
  cwd: frontendDir,
  encoding: "utf8",
})
  .split("\0")
  .filter(
    (file) =>
      (file.startsWith("src/") && /\.(?:js|vue)$/.test(file)) ||
      (file.startsWith("scripts/") && /\.(?:js|mjs)$/.test(file)) ||
      (!file.includes("/") && /\.(?:js|mjs)$/.test(file))
  );

if (files.length === 0) {
  console.log("No frontend source files changed; the full lint report still records existing findings.");
} else {
  const eslint = new ESLint({ cwd: frontendDir });
  const results = await eslint.lintFiles(files);
  const formatter = await eslint.loadFormatter("stylish");

  console.log(await formatter.format(results));
  process.exitCode = results.some((result) => result.errorCount > 0) ? 1 : 0;
}
