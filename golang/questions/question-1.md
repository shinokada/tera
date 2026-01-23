Please follow Claude.md instruction.

# Questions about golang/spec-documents/flow-charts.md

When you change chart, please update other documents/files in golang/spec-documents directory to keep everything consistent and in sync.

- According to Application Overview in golang/spec-documents/flow-charts.md, it has SavePrompt after playback. Since the list for Quickplay is from My favorites list, there is no need to save it again. This applies to the Main Menu Screen as well. Please update all flow charts relating to this, implementation-plan.md, technical-approach.md, keyboard-shortcuts-guide.md, API_SPEC.md, GETTING_STARTED.md, README.md, TESTING.md.

- Use fzf-style display only for ration stations where there are many items.
In Play Screen, it has Display Lists with fzf-style, but there won't be so many list items, so I don't think you need to use fzf-style display. What do you think?

- In Play Screen flow chart, after Playback, there is Save to Quick Favorites and Stop Playback -> Show Save Prompt. What do you think having two points to save to favorites.
