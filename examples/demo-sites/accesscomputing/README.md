# AccessComputing Demo

The University of Washington's AccessComputing program provides this demonstration site for accessibility education.

## URLs

| Version | URL |
|---------|-----|
| Before (Inaccessible) | https://projects.accesscomputing.uw.edu/au/before.html |
| After (Accessible) | https://projects.accesscomputing.uw.edu/au/after.html |

## About This Demo

AccessComputing is a National Science Foundation alliance that promotes the participation of people with disabilities in computing fields. This demo site is part of their educational resources.

## Accessibility Issues Demonstrated

### Before Version

- Missing document language
- Poor heading structure
- Images without alt text
- Inaccessible navigation
- Color-only information
- Missing form labels

### After Version

- Proper `lang` attribute
- Logical heading hierarchy
- Descriptive alt text
- Keyboard-accessible navigation
- Multiple visual cues
- Associated labels for all inputs

## Reports

| Version | JSON | Markdown | PDF |
|---------|------|----------|-----|
| Before | [vpat.json](./before/vpat.json) | [vpat.md](./before/vpat.md) | [vpat.pdf](./before/vpat.pdf) |
| After | [vpat.json](./after/vpat.json) | [vpat.md](./after/vpat.md) | [vpat.pdf](./after/vpat.pdf) |
| Comparison | [comparison.json](./comparison/comparison.json) | [comparison.md](./comparison/comparison.md) | [comparison.pdf](./comparison/comparison.pdf) |

## Generate Reports

```bash
agenta11y compare \
  https://projects.accesscomputing.uw.edu/au/before.html \
  https://projects.accesscomputing.uw.edu/au/after.html \
  --name "AccessComputing Demo" \
  -f markdown -o examples/demo-sites/accesscomputing/comparison/comparison.md
```

## References

- [AccessComputing](https://www.washington.edu/accesscomputing/)
- [AccessComputing Resources](https://www.washington.edu/accesscomputing/resources)
