import json
import os
import sys

def load_results(path):
    if not os.path.exists(path):
        print(f"Warning: Results file not found at {path}")
        return {}
    with open(path, "r") as f:
        data = json.load(f)
    # Map by task name
    return {item["name"]: item for item in data}

def get_category(task_name):
    if "deployment" in task_name or task_name == "create-deployment":
        return "Workload Deployment"
    elif "computeclass" in task_name:
        return "Compute Resource Management"
    elif "gateway" in task_name:
        return "Gateway Ingress & Routing"
    elif "hpa" in task_name:
        return "Custom Metrics Autoscaling"
    return "Other"

def format_percentage(val):
    if val is None:
        return "N/A"
    return f"{int(val * 100)}%"

def get_status_emoji(diff):
    if diff > 0.05:
        return f"🟢 +{int(diff * 100)}%"
    elif diff < -0.05:
        return f"🔴 -{int(abs(diff) * 100)}%"
    return f"⚪ {int(diff * 100)}%"

def get_score(scores, name):
    if not isinstance(scores, dict):
        return None
    # Try exact match
    if name in scores:
        val = scores[name]
        return val.get("score") if isinstance(val, dict) else val
    # Try with GEval suffix
    suffix_name = f"{name} [GEval]"
    if suffix_name in scores:
        val = scores[suffix_name]
        return val.get("score") if isinstance(val, dict) else val
    return None

def main():
    base_path = "devops-bench/results/base/results.json"
    enhanced_path = "devops-bench/results/enhanced/results.json"
    
    if len(sys.argv) > 2:
        base_path = sys.argv[1]
        enhanced_path = sys.argv[2]

    base_data = load_results(base_path)
    enhanced_data = load_results(enhanced_path)

    if not base_data and not enhanced_data:
        print("Error: No evaluation results found to summarize.")
        sys.exit(1)

    all_tasks = sorted(list(set(base_data.keys()) | set(enhanced_data.keys())))
    
    # Categories definition
    categories = [
        "Workload Deployment",
        "Compute Resource Management",
        "Gateway Ingress & Routing",
        "Custom Metrics Autoscaling"
    ]
    
    # Aggregators: {category: {"base_checklist": [], "enhanced_checklist": [], "base_outcome": [], "enhanced_outcome": []}}
    agg = {cat: {"bc": [], "ec": [], "bo": [], "eo": []} for cat in categories}
    
    detailed_rows = []
    failures = []

    for task in all_tasks:
        b_task = base_data.get(task, {})
        e_task = enhanced_data.get(task, {})

        cat = get_category(task)
        
        # Extract scores
        b_scores = b_task.get("scores", {})
        e_scores = e_task.get("scores", {})

        b_checklist = get_score(b_scores, "ChecklistScore")
        e_checklist = get_score(e_scores, "ChecklistScore")
        b_outcome = get_score(b_scores, "OutcomeValidity")
        e_outcome = get_score(e_scores, "OutcomeValidity")

        b_lat = b_task.get("latency", 0.0)
        e_lat = e_task.get("latency", 0.0)

        # Aggregate for category average
        if cat in agg:
            if b_checklist is not None: agg[cat]["bc"].append(b_checklist)
            if e_checklist is not None: agg[cat]["ec"].append(e_checklist)
            if b_outcome is not None: agg[cat]["bo"].append(b_outcome)
            if e_outcome is not None: agg[cat]["eo"].append(e_outcome)

        # Build detailed matrix row
        detailed_rows.append(
            f"| `{task}` | {format_percentage(b_checklist)} | **{format_percentage(e_checklist)}** | {format_percentage(b_outcome)} | **{format_percentage(e_outcome)}** | {e_lat:.1f}s |"
        )

        # Check for checklist failures in Enhanced Agent to highlight
        checklist_data = e_scores.get("ChecklistScore", {})
        if not checklist_data.get("success", True) or (e_checklist is not None and e_checklist < 1.0):
            # Collect failures reasons
            fail_items = []
            for score_name, score_val in e_scores.items():
                if score_name.startswith("Check: ") and not score_val.get("success", True):
                    fail_items.append(score_name[7:])
            
            reasons_str = "<br>".join([f"• {item}" for item in fail_items]) if fail_items else checklist_data.get("reason", "Unknown checklist gap")
            failures.append((task, reasons_str))

    # Calculate Category Averages
    cat_rows = []
    for cat in categories:
        b_c_avg = sum(agg[cat]["bc"]) / len(agg[cat]["bc"]) if agg[cat]["bc"] else None
        e_c_avg = sum(agg[cat]["ec"]) / len(agg[cat]["ec"]) if agg[cat]["ec"] else None
        b_o_avg = sum(agg[cat]["bo"]) / len(agg[cat]["bo"]) if agg[cat]["bo"] else None
        e_o_avg = sum(agg[cat]["eo"]) / len(agg[cat]["eo"]) if agg[cat]["eo"] else None

        delta = (e_c_avg - b_c_avg) if (e_c_avg is not None and b_c_avg is not None) else 0.0
        
        cat_rows.append(
            f"| **{cat}** | {format_percentage(b_c_avg)} | **{format_percentage(e_c_avg)}** | {get_status_emoji(delta)} | {format_percentage(b_o_avg)} | **{format_percentage(e_o_avg)}** |"
        )

    # Compile the final Markdown report
    markdown = f"""# 🤖 GKE-MCP DevOps Bench Daily Evaluation Report

This report summarizes the daily performance comparison between the **GKE-MCP Enhanced Agent** and the **Baseline Gemini Model** (MCP disabled) on GKE DevOps tasks.

## 📊 GKE Category Averages (Accuracy & Compliance)

| Category | Base Checklist | Enhanced Checklist | Checklist Delta | Base Outcome | Enhanced Outcome |
| :--- | :---: | :---: | :---: | :---: | :---: |
{chr(10).join(cat_rows)}

---

## 🎯 Detailed Task Matrix

| Task Name | Base Checklist | Enhanced Checklist | Base Outcome | Enhanced Outcome | Agent Latency |
| :--- | :---: | :---: | :---: | :---: | :---: |
{chr(10).join(detailed_rows)}

---

"""

    if failures:
        fail_sections = []
        for task, reason in failures:
            fail_sections.append(f"""### 🔍 `{task}` Checklist Gaps
> [!WARNING]
> **Failed Checklist Items:**
> {reason}
""")
        
        markdown += f"""## ⚠️ Identified Checklist Gaps (Enhanced Agent Gaps)

<details>
<summary>Click to view details of failed checks per task</summary>

{"".join(fail_sections)}
</details>

---
"""

    markdown += f"""> [!NOTE]
> This evaluation was run automatically on Gactions Runner. Results will sync to Confident AI once credentials are provided.
"""

    # Write report
    summary_env = os.getenv("GITHUB_STEP_SUMMARY")
    if summary_env:
        with open(summary_env, "w") as f:
            f.write(markdown)
        print("Successfully wrote Job Summary to $GITHUB_STEP_SUMMARY")
    else:
        # Local print for debugging
        print(markdown)

if __name__ == "__main__":
    main()
