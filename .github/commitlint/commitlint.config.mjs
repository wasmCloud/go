// commitlint configuration: https://commitlint.js.org/reference/configuration.html
// Checked by .github/workflows/commitlint.yml.
import conventional from '@commitlint/config-conventional';

const [, , types] = conventional.rules['type-enum'];

export default {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // A subject line must be under 70 characters.
    'header-max-length': [2, 'always', 69],
    // release-train.yaml commits `release: component/vX.Y.Z`.
    'type-enum': [2, 'always', [...types, 'release']],
    // Dependabot writes `Bump x from a to b`.
    'subject-case': [0],
    // Bodies hold links and generated text, such as Dependabot's release notes.
    'body-max-line-length': [0],
    'footer-max-line-length': [0],
  },
};
