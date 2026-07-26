module.exports = {
    apps: [
        // autop:begin autop-agy-system
        {
            name: "AutoP agy system",
            script: "autop",
            args: ["-c", "agy", "-t", "system"],
            cwd: "/Users/shuk/projects/cc-plugin",
            namespace: "autop",
            instances: 1,
            optional: true,
            cron: "10 0-9 * * *",
            autorestart: false,
            watch: false
        },
        // autop:end autop-agy-system
        // autop:begin autop-agy-auto-evolving
        {
            name: "AutoP agy auto-evolving",
            script: "autop",
            args: ["-c", "agy", "-t", "auto-evolving", "--bypass-permission=true", "--model", "gemini-3.6-flash-medium", "--effort", "low"],
            cwd: "/Users/shuk/projects/cc-plugin",
            namespace: "autop",
            instances: 1,
            optional: true,
            cron: "8 6 * * *",
            autorestart: false,
            watch: false
        },
        // autop:end autop-agy-auto-evolving
        // autop:begin autop-claudem-auto-evolving
        {
            name: "AutoP claudem auto-evolving",
            script: "autop",
            args: ["-c", "claudem", "-t", "auto-evolving", "--bypass-permission=true", "--model", "MiniMax-M3", "--effort", "max"],
            cwd: "/Users/shuk/projects/cc-plugin",
            namespace: "autop",
            instances: 1,
            optional: true,
            cron: "44 3 * * *",
            autorestart: false,
            watch: false
        },
        // autop:end autop-claudem-auto-evolving
        // autop:begin autop-codex-auto-evolving
        {
            name: "AutoP codex auto-evolving",
            script: "autop",
            args: ["-c", "codex", "-t", "auto-evolving", "--bypass-permission=true", "--model", "gpt-5.5", "--effort", "max"],
            cwd: "/Users/shuk/projects/cc-plugin",
            namespace: "autop",
            instances: 1,
            optional: true,
            cron: "12 5 * * *",
            autorestart: false,
            watch: false
        },
        // autop:end autop-codex-auto-evolving
    ]
};
