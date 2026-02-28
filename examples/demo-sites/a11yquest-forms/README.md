# A11yQuest Forms Demo

A11yQuest provides interactive accessibility demonstrations, including this forms-focused demo that illustrates common form accessibility issues and their fixes.

## URLs

| Version | URL |
|---------|-----|
| Before (Inaccessible) | https://www.a11yquest.com/demos/forms/before |
| After (Accessible) | https://www.a11yquest.com/demos/forms/after |

## About This Demo

A11yQuest offers hands-on accessibility learning through practical examples. The forms demo focuses on one of the most common accessibility problem areas: form inputs and their labels.

## Accessibility Issues Demonstrated

### Before Version

- Input fields without associated labels
- Placeholder text used as labels
- Missing fieldset/legend for grouped inputs
- No error identification
- Color-only error indication
- Inaccessible custom dropdowns
- Missing autocomplete attributes

### After Version

- Properly associated `<label>` elements
- Visible labels (placeholders as hints only)
- Fieldset/legend for radio groups and checkboxes
- Programmatic error identification with `aria-describedby`
- Multiple error indicators (color, icon, text)
- Accessible custom components with ARIA
- Appropriate autocomplete attributes

## WCAG Criteria Covered

| Criterion | Level | Description |
|-----------|-------|-------------|
| 1.3.1 | A | Info and Relationships |
| 1.3.5 | AA | Identify Input Purpose |
| 2.4.6 | AA | Headings and Labels |
| 3.3.1 | A | Error Identification |
| 3.3.2 | A | Labels or Instructions |
| 3.3.3 | AA | Error Suggestion |
| 4.1.2 | A | Name, Role, Value |

## Reports

| Version | JSON | Markdown | PDF |
|---------|------|----------|-----|
| Before | [vpat.json](./before/vpat.json) | [vpat.md](./before/vpat.md) | [vpat.pdf](./before/vpat.pdf) |
| After | [vpat.json](./after/vpat.json) | [vpat.md](./after/vpat.md) | [vpat.pdf](./after/vpat.pdf) |
| Comparison | [comparison.json](./comparison/comparison.json) | [comparison.md](./comparison/comparison.md) | [comparison.pdf](./comparison/comparison.pdf) |

## Generate Reports

```bash
# Generate comparison report
agenta11y compare \
  https://www.a11yquest.com/demos/forms/before \
  https://www.a11yquest.com/demos/forms/after \
  --name "A11yQuest Forms Demo" \
  -f markdown -o examples/demo-sites/a11yquest-forms/comparison/comparison.md

# Generate individual VPATs
agenta11y audit https://www.a11yquest.com/demos/forms/before \
  -f vpat -o examples/demo-sites/a11yquest-forms/before/vpat.md

agenta11y audit https://www.a11yquest.com/demos/forms/after \
  -f vpat -o examples/demo-sites/a11yquest-forms/after/vpat.md
```

## References

- [A11yQuest Demos](https://www.a11yquest.com/demos/)
- [WebAIM Forms Tutorial](https://webaim.org/techniques/forms/)
- [WCAG Form Labels](https://www.w3.org/WAI/tutorials/forms/labels/)
