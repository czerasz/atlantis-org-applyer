// Needed variables in GtiLab CI are: GITHUB_TOKEN
// https://github.com/semantic-release/semantic-release/blob/master/docs/usage/ci-configuration.md#authentication

const commitAnalyzerOptions = {
  preset: 'angular',
  releaseRules: [
    { type: 'breaking', release: 'major' },
    { type: 'refactor', release: 'patch' },
    { type: 'config', release: 'patch' },
    { scope: 'no-release', release: false },
    { scope: 'test', release: false },
  ],
  parserOpts: {
    noteKeywords: [],
  },
};

const releaseNotesGeneratorOptions = {
  writerOpts: {
    transform: (commit, context) => {
      const issues = [];
      const types = {
        feat: 'Features',
        fix: 'Bug Fixes',
        style: 'Styles',
        refactor: 'Code Refactoring',
        test: 'Tests',
        build: 'Builds',
        ci: 'Continuous Integrations',
        perf: 'Performance Improvements',
        docs: 'Documentation',
        chore: 'Chores',
        revert: 'Reverts',
        breaking: 'Breaking',
        config: 'Configuration',
      }

      if (commit.type === 'no-release') {
        return;
      }

      if (commit.type in types) {
        commit.type = types[commit.type];
      }

      if (typeof commit.hash === 'string') {
        commit.shortHash = commit.hash.substring(0, 7);
      }

      if (typeof commit.subject === 'string') {
        let url = context.repository
          ? `${context.host}/${context.owner}/${context.repository}`
          : context.repoUrl;
        if (url) {
          url = `${url}/issues/`;
          // Issue URLs.
          commit.subject = commit.subject.replace(/#([0-9]+)/g, (_, issue) => {
            issues.push(issue);
            return `[#${issue}](${url}${issue})`;
          });
        }
        if (context.host) {
          // User URLs.
          commit.subject = commit.subject.replace(
            /\B@([a-z0-9](?:-?[a-z0-9/]){0,38})/g,
            (_, username) => {
              if (username.includes('/')) {
                return `@${username}`;
              }

              return `[@${username}](${context.host}/${username})`;
            }
          );
        }
      }

      // remove references that already appear in the subject
      commit.references = commit.references.filter((reference) => {
        if (issues.indexOf(reference.issue) === -1) {
          return true;
        }

        return false;
      });

      return commit;
    },
  },
};

module.exports = {
  branches: ["main"],
  plugins: [
    // analyze commits with conventional-changelog
    ['@semantic-release/commit-analyzer', commitAnalyzerOptions],
    // generate changelog content with conventional-changelog
    ['@semantic-release/release-notes-generator', releaseNotesGeneratorOptions],
    // updates the changelog file
    "@semantic-release/changelog",
    // creating a git tag
    '@semantic-release/github',
    // creating a new version commit
    ["@semantic-release/git", {
      "assets": ["CHANGELOG.md"],
      "message": "chore(release): ${nextRelease.version} \n\n${nextRelease.notes}"
    }],
  ],
};
