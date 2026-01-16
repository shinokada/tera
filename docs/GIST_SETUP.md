# Setting Up GitHub Gist Integration

TERA uses GitHub Gists to backup and restore your favorite radio station lists. To use this feature, you need to set up a GitHub Personal Access Token.

## Quick Setup

1. **Copy the environment template:**
   ```bash
   cp .env.example .env
   ```

2. **Create a GitHub Personal Access Token:**
   - Go to [GitHub Token Settings](https://github.com/settings/tokens)
   - Click "Generate new token (classic)"
   - Give it a name like "TERA Gist Access"
   - Select **only** the `gist` scope
   - Click "Generate token"
   - Copy the token (you won't be able to see it again!)

3. **Add your token to the .env file:**
   ```bash
   # Edit .env file
   nano .env  # or use your preferred editor
   
   # Replace 'your_github_token_here' with your actual token
   GITHUB_TOKEN=ghp_YourActualTokenHere123456789
   ```

4. **Save and you're done!**
   The token will be automatically loaded when you run TERA.

## Security Notes

- ⚠️ **Never commit the `.env` file to git** - it's already in `.gitignore`
- ⚠️ **Keep your token secret** - don't share it with anyone
- ℹ️ The token is only stored locally on your machine
- ℹ️ Only the `gist` scope is needed - don't add unnecessary permissions

## Troubleshooting

If you get a "404 Not Found" error:
- Make sure your token is correctly copied to `.env`
- Verify the token has the `gist` scope enabled
- Try regenerating the token if it's expired

If you get authentication errors:
- Check that your `.env` file is in the root of the tera directory
- Make sure there are no extra spaces around the token
- Ensure the file is named exactly `.env` (not `.env.txt`)

## Using Gist Features

Once set up, you can:
1. **Create a gist**: Uploads all your station lists to a secret GitHub Gist
2. **Recover from gist**: Download and restore your lists from a Gist URL
