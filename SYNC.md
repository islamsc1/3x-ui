# AllowInsecure Custom Patch Documentation

This repository maintains a custom patch that permanently restores the `allowInsecure` option for self-managed/self-signed certificates.

## Reversion Context
- **Feature**: `allowInsecure` (skip TLS verification on client connections)
- **Reverted Upstream Commit**: [511adffc5bb419a938e3b1eb0087a49f62d76db6](https://github.com/MHSanaei/3x-ui/commit/511adffc5bb419a938e3b1eb0087a49f62d76db6) ("Remove allowInsecure")
- **Rationale for Restore**: Enables self-signed, private, or expired certificate usage.
- **Security Tradeoff**: **WARNING!** Enabling this feature disables TLS peer certificate checks on client machines. Clients are vulnerable to Man-in-the-Middle (MitM) attacks. Enable only on trusted private networks or with full understanding of the risks.

---

## Upgrade & Maintenance Workflow

When a new upstream release is published, follow these steps to keep the fork current and merge the changes cleanly.

### Step 1: Fetch Latest Upstream Code
```bash
git fetch upstream --tags
```

### Step 2: Checkout and Rebase the Custom Branch
Rebase our long-lived `with-allow-insecure` branch onto the new release tag (e.g., `v3.4.0`):
```bash
git checkout with-allow-insecure
git rebase v3.4.0
```

### Step 3: Resolve Merge Conflicts (If Any)
If upstream modified any of the subscription or frontend files we patch:
1. Locate the conflicting blocks.
2. Resolve conflicts, ensuring both upstream changes and our `allowInsecure` wiring coexist.
3. Mark conflicts resolved and continue:
   ```bash
   git add <conflict-files>
   git rebase --continue
   ```

### Step 4: Verify and Build
1. **Backend**: Build the Go binary to verify there are no syntax or type errors:
   ```bash
   go build
   ```
2. **Frontend**: Compile the React/TypeScript bundle to ensure compilation succeeds:
   ```bash
   cd frontend
   npm run build
   ```

### Step 5: Push Updates
Force-push the rebased branch to your remote fork:
```bash
git push origin with-allow-insecure --force
```
