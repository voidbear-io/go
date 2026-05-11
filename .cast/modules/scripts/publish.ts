import { format, tryParse } from "jsr:@std/semver@1";
import { join, resolve } from "jsr:@frostyeti/path@0.0.0-alpha.0";
import { existsSync } from "jsr:@frostyeti/fs@0.0.0-alpha.0";

function usage(): never {
    console.error(`Usage: publish.ts <mod> [options]

Options:
  Version bump (requires --bump for increment, or set with --value when not bumping):
    --bump
    --patch | -p    (--value <n> when used without --bump)
    --minor | -m
    --major | -M
    --value <semver>

  Prerelease / metadata:
    --release         Drop prerelease and build metadata (keep major.minor.patch only)
    --alpha [n]       Set prerelease to alpha.n (default n=0), e.g. --alpha 5 -> -alpha.5
    --beta [n]        Set prerelease to beta.n (default n=0)
    --rc [n]          Set prerelease to rc.n (default n=0)
    (at most one of --alpha, --beta, --rc; cannot combine with --release)

  Other:
    --tag | -t        Create git annotated tag for the new version
    --skip-push       With --tag: create the tag locally but do not git push --tags`);
    Deno.exit(1);
}

function parsePrereleaseFlag(
    args: string[],
    flag: string,
): { found: boolean; num: number } {
    const idx = args.indexOf(flag);
    if (idx === -1) return { found: false, num: 0 };
    const next = args[idx + 1];
    if (
        next !== undefined && /^[0-9]+$/.test(next)
    ) {
        return { found: true, num: Number.parseInt(next, 10) };
    }
    return { found: true, num: 0 };
}

const args = Deno.args;
if (args.length === 0) {
    usage();
}

const dir = import.meta.dirname;
if (dir === undefined) {
    console.error("import.meta.dirname is not available");
    Deno.exit(1);
}
const rootDir = resolve(join(dir, "..", "..", ".."));
const path = `${rootDir}/.versions.json`;

if (existsSync(path) === false) {
    console.error(`.versions.json file not found at path: ${path}`);
    Deno.exit(1);
}

const versions = JSON.parse(Deno.readTextFileSync(path));
const mod = args[0];

if (!mod || mod.startsWith("-")) {
    usage();
}

if (existsSync(join(rootDir, mod)) === false) {
    console.error(`Module directory not found: ${mod}`);
    Deno.exit(1);
}

const bump = args.includes("--bump");
const patch = args.includes("--patch") || args.includes("-p");
const minor = args.includes("--minor") || args.includes("-m");
const major = args.includes("--major") || args.includes("-M");
const tag = args.includes("--tag") || args.includes("-t");
const skipPush = args.includes("--skip-push");
const release = args.includes("--release");

if (skipPush && !tag) {
    console.error("--skip-push only applies with --tag");
    Deno.exit(1);
}

const alpha = parsePrereleaseFlag(args, "--alpha");
const beta = parsePrereleaseFlag(args, "--beta");
const rc = parsePrereleaseFlag(args, "--rc");

const preCount = [alpha, beta, rc].filter((p) => p.found).length;
if (preCount > 1) {
    console.error("At most one of --alpha, --beta, --rc may be set");
    Deno.exit(1);
}
if (release && preCount > 0) {
    console.error("--release cannot be combined with --alpha, --beta, or --rc");
    Deno.exit(1);
}

const setValue = args.indexOf("--value");
let value: string | undefined = undefined;
if (setValue > -1) {
    if (setValue + 1 >= args.length) {
        console.error("Usage: publish.ts <mod> --value <new_version>");
        Deno.exit(1);
    }
    value = args[setValue + 1];
}

const ver = versions[mod];
let version = tryParse(ver);
if (version === null) {
    console.error(`Invalid semantic version for module ${mod}: ${ver}`);
    Deno.exit(1);
}

if (bump) {
    if (patch) {
        version!.patch++;
    } else if (minor) {
        version!.minor++;
        version!.patch = 0;
    } else if (major) {
        version!.major++;
        version!.minor = 0;
        version!.patch = 0;
    }
} else {
    if (patch) {
        const next = Number.parseInt(value!);
        if (isNaN(next)) {
            console.error("Invalid value for patch version");
            Deno.exit(1);
        }
        version!.patch = next;
    } else if (minor) {
        const next = Number.parseInt(value!);
        if (isNaN(next)) {
            console.error("Invalid value for minor version");
            Deno.exit(1);
        }
        version!.minor = next;
        version!.patch = 0;
    } else if (major) {
        const next = Number.parseInt(value!);
        if (isNaN(next)) {
            console.error("Invalid value for major version");
            Deno.exit(1);
        }
        version!.major = next;
        version!.minor = 0;
        version!.patch = 0;
    } else if (value) {
        const newVersion = tryParse(value);
        if (newVersion === null) {
            console.error(`Invalid semantic version: ${value}`);
            Deno.exit(1);
        }
        console.log(`Setting version to ${newVersion}`);
        version = newVersion;
    }
}

if (release) {
    version!.prerelease = [];
    version!.build = [];
}

if (alpha.found) {
    version!.prerelease = ["alpha", alpha.num];
    version!.build = [];
} else if (beta.found) {
    version!.prerelease = ["beta", beta.num];
    version!.build = [];
} else if (rc.found) {
    version!.prerelease = ["rc", rc.num];
    version!.build = [];
}

versions[mod] = format(version!);
Deno.writeTextFileSync(path, JSON.stringify(versions, null, 4));
console.log(`${mod} -> ${versions[mod]}`);

if (tag) {
    if (bump) {
        const cmd = new Deno.Command("git", {
            args: ["commit", "-a", "-m", "update module version for " + mod],
            cwd: dir,
            stdout: "inherit",
            stderr: "inherit",
        });

        const result = await cmd.output();
        if (result.code !== 0) {
            console.error(`Failed to tag version for ${mod}`);
            Deno.exit(result.code);
        }
    }

    let cmd = new Deno.Command("git", {
        args: [
            "tag",
            "-a",
            `${mod}/v${versions[mod]}`,
            "-m",
            `Release v${versions[mod]}`,
        ],
        cwd: dir,
        stdout: "inherit",
        stderr: "inherit",
    });

    const tagResult = await cmd.output();
    if (tagResult.code !== 0) {
        console.error(`Failed to create tag for ${mod}`);
        Deno.exit(tagResult.code);
    }

    if (!skipPush) {
        cmd = new Deno.Command("git", {
            args: ["push", "--tags"],
            cwd: dir,
            stdout: "inherit",
            stderr: "inherit",
        });

        const pushResult = await cmd.output();
        if (pushResult.code !== 0) {
            console.error(`Failed to push tags for ${mod}`);
            Deno.exit(pushResult.code);
        }
    }
}
