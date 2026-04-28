# Standard Skill Structure

```text
skills/<skill-name>/
├── SKILL.md               # Core diagnostic workflow
├── README.md              # Public overview
├── BUILD                  # Build definition
├── EVAL.textproto         # Evaluation suite
├── TEST.md                # Manual test plan
├── references/
│   └── failure_signatures.md  # Authentic log/metric examples
└── scripts/
    └── validate_queries.sh    # Syntax validator for filters
```
