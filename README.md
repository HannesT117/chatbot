# claude template

This repository contains a few claude configurations that I find helpful or want to try.

My local setup and decision flow for when to use which model is described in the [docs](docs/local-setup.md).

## Getting started

1. Create a new repository with this template: `gh repo create <my-project> -p HannesT117/claude-template -l MIT --public -c`

2. Search for occurences of `CLAUDE_SETUP` and add project specific information.

3. Check the plugins in [.claude/settings.json](.claude/settings.json) for unnecessary plugins and remove them.

## Remarks

### Update Claude.md regularly

End corrections with: "Now update CLAUDE.md so you don't make that mistake again."

Keep iterating until the mistake rate measurably drops.

## Sources and Inspiration

- [Boris Cherny's Claude Config](https://github.com/0xquinto/bcherny-claude/)
- [SkillsMP, Marketplace for AI agent skills](https://skillsmp.com/)
- [Which Agent, helper to figure out which tool to use](https://hannest117.github.io/which-agent/)
- [llmfit, find the best llm for your machine](https://github.com/AlexsJones/llmfit)
