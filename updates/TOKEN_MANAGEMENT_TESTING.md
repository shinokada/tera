# Token Management - Testing Checklist

## Pre-Implementation Testing

- [x] Syntax validation (bash -n) for all modified scripts
- [x] No breaking changes to existing functionality

---

## Core Functionality Testing

### Token Storage Functions

- [ ] `init_token_directory()` - Creates ~/.config/tera/tokens/ with 700 permissions
- [ ] `save_github_token()` - Saves token file with 600 permissions
- [ ] `load_github_token()` - Returns token from file
- [ ] `has_github_token()` - Returns true/false correctly
- [ ] `delete_github_token()` - Removes token file
- [ ] `get_masked_token()` - Returns proper format (ghp_...xyz)
- [ ] `validate_github_token()` - Tests with GitHub API

### Menu Navigation

- [ ] Main Menu → \"6) Gist\" works
- [ ] Gist Menu displays all options including \"1) Token Management\"
- [ ] Token Management Menu displays all 4 options
- [ ] Navigation back works (0 to go back)
- [ ] All menus use fzf selection properly

---

## Setup Token Workflow

### Happy Path (Valid Token)

- [ ] Can access \"1) Setup/Change Token\"
- [ ] Prompts for token with hidden input
- [ ] Validates token format (not too short, has prefix)
- [ ] Makes API call to GitHub
- [ ] Shows GitHub username on success
- [ ] Saves to ~/.config/tera/tokens/github_token
- [ ] Returns to Token Management menu
- [ ] Token can be used in Gist operations

### Format Validation

- [ ] Rejects empty token
- [ ] Rejects too-short token (< 20 chars)
- [ ] Accepts token starting with \"ghp_\"
- [ ] Offers retry on validation failure

### Replacement Flow

- [ ] If token exists, asks \"Replace existing token?\"
- [ ] Allows user to cancel replacement
- [ ] If user accepts, replaces token
- [ ] New token is validated before saving

### Error Handling

- [ ] Network error shows helpful message
- [ ] Invalid token shows specific error
- [ ] Offers to retry on failure
- [ ] Graceful fallback to menu on cancel

---

## View Token Workflow

- [ ] Can access \"2) View Current Token\"
- [ ] Shows masked token (ghp_...xyz)
- [ ] Shows GitHub username
- [ ] Validates token before showing info
- [ ] Shows \"✓ Token is currently valid\" on success
- [ ] Shows warning on validation failure
- [ ] Returns to menu on Enter

---

## Validate Token Workflow

### Valid Token

- [ ] Can access \"3) Validate Token\"
- [ ] Shows \"✓ Token is VALID!\"
- [ ] Displays GitHub username
- [ ] Shows confirmation message
- [ ] Returns to menu

### Invalid Token

- [ ] Shows \"✗ Token is INVALID or EXPIRED!\"
- [ ] Lists possible causes
- [ ] Suggests updating token
- [ ] Returns to menu

---

## Delete Token Workflow

### Confirmation Flow

- [ ] Can access \"4) Delete Token\"
- [ ] Shows warning with masked token
- [ ] Prompts for \"yes\" confirmation
- [ ] Requires exact \"yes\" (not just any input)
- [ ] Cancels on any other input

### Deletion

- [ ] Shows \"✓ Token has been deleted successfully!\"
- [ ] Removes token file
- [ ] Unsets GITHUB_TOKEN environment variable
- [ ] Future gist operations show \"token not found\"
- [ ] Can set up new token anytime

---

## Integration with Gist Operations

### Create Gist (with token)

- [ ] Token is loaded from secure storage
- [ ] Token is used in API request
- [ ] Gist creation succeeds
- [ ] Gist metadata is saved

### Create Gist (without token)

- [ ] Shows helpful error message
- [ ] Directs user to Token Management
- [ ] Doesn't create gist
- [ ] Returns to menu

### Other Gist Operations

- [ ] My Gists works with token
- [ ] Recover Gist works with token
- [ ] Update Gist works with token
- [ ] Delete Gist works with token
- [ ] All show error without token

---

## Documentation Testing

- [ ] GIST_SETUP.md is updated and clear
- [ ] TOKEN_MANAGEMENT.md is comprehensive
- [ ] TOKEN_MANAGEMENT_VISUAL_GUIDE.md is accurate
- [ ] README.md references token system
- [ ] All links work
- [ ] Examples are accurate

---

## User Experience Testing

### First-Time User

- [ ] Can complete setup without reading docs
- [ ] Clear prompts guide through process
- [ ] Menu structure is intuitive
- [ ] Feedback is immediate
- [ ] Can recover from mistakes

### Returning User

- [ ] Token is remembered between sessions
- [ ] No need to set up again
- [ ] Can easily change token
- [ ] Token status is visible

### Error Recovery

- [ ] Can retry on validation failure
- [ ] Can cancel and start over
- [ ] Can delete and redo
- [ ] Clear error messages
- [ ] Helpful troubleshooting info

---

## Security Testing

### File Permissions

- [ ] Token file is created with 600 permissions
- [ ] Token directory has 700 permissions
- [ ] Only owner can read/write token
- [ ] Permissions persist after restart

### Input Security

- [ ] Token input is hidden (no echo)
- [ ] Masked display shows only safe portion
- [ ] Token not logged in error messages
- [ ] Token not shown in process list

### API Validation

- [ ] Actually calls GitHub API
- [ ] Doesn't save invalid tokens
- [ ] Validates before saving
- [ ] Shows user feedback

---

## Edge Cases

- [ ] Multiple rapid setup attempts
- [ ] Rapid cancel/retry cycles
- [ ] Very long token (copy extra characters)
- [ ] Token with whitespace
- [ ] Expired token (revoked on GitHub)
- [ ] Network timeout during validation
- [ ] Permission issues on token directory
- [ ] Token file corrupted (invalid JSON)
- [ ] Concurrent TERA instances

---

## Performance Testing

- [ ] Token loading is instant
- [ ] Menu navigation is responsive
- [ ] API validation completes in reasonable time (< 5s)
- [ ] No slowdown of other TERA operations
- [ ] Gist operations still work at same speed

---

## Installation Method Testing

Test with all installation methods:

- [ ] **Source:** `./tera` from directory
- [ ] **Brew:** `brew install tera` (if available)
- [ ] **.deb:** `apt install tera` (if available)
- [ ] **/awesome:** `awesome install tera` (if available)

For each:
- [ ] Can set up token
- [ ] Token persists
- [ ] Gist operations work
- [ ] Multiple sessions work

---

## Platform Testing

- [ ] macOS (BSD date)
- [ ] Linux (GNU date)
- [ ] Different shells (bash, zsh, sh)
- [ ] Different terminal emulators

---

## Final Checklist

- [ ] All syntax valid
- [ ] All functions work independently
- [ ] All workflows complete successfully
- [ ] Gist operations still work
- [ ] No breaking changes
- [ ] Documentation is complete
- [ ] User experience is smooth
- [ ] Security is maintained
- [ ] Error messages are helpful
- [ ] Menu is intuitive
- [ ] Works on all platforms
- [ ] Works with all installations
- [ ] Ready for production

---

## Post-Implementation

- [ ] Commit changes to git
- [ ] Update CHANGELOG
- [ ] Create release notes
- [ ] Tag version
- [ ] Update package managers
- [ ] Announce to users
- [ ] Monitor for issues
- [ ] Update wiki/docs as needed

---

## Notes

```
Key Points:

1. Token storage at ~/.config/tera/tokens/github_token
2. File permissions must be 600 (user only)
3. Token loaded at TERA startup
4.
5. All validations happen before saving
6. Masked display: ghp_VaVnzkTqr...ItDAAEo
7. No encryption (relies on OS permissions)
8. Works with all installation methods
9. Complete documentation provided
```
