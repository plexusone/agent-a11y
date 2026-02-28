# W3C BAD (Before-After Demonstration)

The W3C Web Accessibility Initiative (WAI) maintains this official demonstration site to showcase accessibility issues and their remediation.

## URLs

| Version | URL |
|---------|-----|
| Before (Inaccessible) | https://www.w3.org/WAI/demos/bad/before/home.html |
| After (Accessible) | https://www.w3.org/WAI/demos/bad/after/home.html |

## About This Demo

The BAD (Before-and-After Demonstration) was created by W3C WAI to:

- Illustrate common accessibility barriers
- Demonstrate how to repair accessibility issues
- Provide training material for web developers

### Pages Included

The demo site includes several pages demonstrating different accessibility scenarios:

1. **Home** - General layout and navigation issues
2. **News** - Content structure and headings
3. **Tickets** - Forms and input labels
4. **Survey** - Complex form interactions
5. **Template** - Page templates and consistency

## Accessibility Issues Demonstrated

### Before Version (Inaccessible)

Common issues found in the "before" version:

- Missing alt text on images
- Poor color contrast
- Missing form labels
- Improper heading structure
- Keyboard navigation issues
- Missing skip links
- Inaccessible tables

### After Version (Accessible)

The "after" version demonstrates fixes including:

- Descriptive alt text for all images
- Sufficient color contrast (4.5:1 minimum)
- Properly associated form labels
- Logical heading hierarchy
- Full keyboard accessibility
- Skip navigation links
- Properly structured data tables

## Reports

| Version | JSON | Markdown | PDF |
|---------|------|----------|-----|
| Before | [vpat.json](./before/vpat.json) | [vpat.md](./before/vpat.md) | [vpat.pdf](./before/vpat.pdf) |
| After | [vpat.json](./after/vpat.json) | [vpat.md](./after/vpat.md) | [vpat.pdf](./after/vpat.pdf) |
| Comparison | [comparison.json](./comparison/comparison.json) | [comparison.md](./comparison/comparison.md) | [comparison.pdf](./comparison/comparison.pdf) |

## Generate Reports

```bash
# Generate all W3C BAD reports
make demo-w3c-bad

# Or individually:
agenta11y audit https://www.w3.org/WAI/demos/bad/before/home.html \
  -f vpat -o examples/demo-sites/w3c-bad/before/vpat.md

agenta11y compare \
  https://www.w3.org/WAI/demos/bad/before/home.html \
  https://www.w3.org/WAI/demos/bad/after/home.html \
  --name "W3C BAD Demo" \
  -f markdown -o examples/demo-sites/w3c-bad/comparison/comparison.md
```

## References

- [W3C BAD Overview](https://www.w3.org/WAI/demos/bad/Overview.html)
- [WAI Tutorials](https://www.w3.org/WAI/tutorials/)
- [WCAG 2.2 Quick Reference](https://www.w3.org/WAI/WCAG22/quickref/)
