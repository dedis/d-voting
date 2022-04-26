import { RankQuestion } from 'types/configuration';

// Sum the position for each candidate such that a low score is better
// (e.g if choice 1 ranked first and then fourth then it will have a
// score of (1-1 + 4-1) = 3) and returns the counts and the candidate(s)
// with the lowest score
const countRankResult = (rankResult: number[][], rank: RankQuestion) => {
  const resultsInPercent: string[] = [];
  const minIndices: number[] = [];
  // the maximum score achievable is (number of choices - 1) * number of ballots
  let min = (rank.Choices.length - 1) * rankResult.length;

  const results = rankResult.reduce((a, b) => {
    return a.map((value, index) => {
      return value + b[index];
    });
  });

  const total = results.reduceRight((a, b) => {
    min = a < min ? a : min;
    min = b < min ? b : min;

    return a + b;
  });

  results.forEach((count, index) => {
    if (count === min) {
      minIndices.push(index);
    }

    const percentage = (1 - count / total) * 100;
    const roundedPercentage = (Math.round(percentage * 100) / 100).toString();
    resultsInPercent.push(roundedPercentage);
  });

  return { resultsInPercent, minIndices };
};

// Count the number of vote for a candidate and returns the counts and
// which candidate(s) in the select.Choices has the most votes
const countSelectResult = (selectResult: number[][]) => {
  const resultsInPercent: string[] = [];
  const maxIndices: number[] = [];
  let max = 0;

  const results = selectResult.reduce((a, b) => {
    return a.map((value, index) => {
      const current = value + b[index];

      if (current >= max) {
        max = current;
      }
      return current;
    });
  });

  results.forEach((count, index) => {
    if (count === max) {
      maxIndices.push(index);
    }

    const percentage = (count / selectResult.length) * 100;
    const roundedPercentage = (Math.round(percentage * 100) / 100).toFixed(2);
    resultsInPercent.push(roundedPercentage);
  });
  return { resultsInPercent, maxIndices };
};

// Count the number of votes for each candidate and returns the counts and the
// candidate(s) with the most votes
const countTextResult = (textResult: string[][]) => {
  const resultsInPercent: Map<string, string> = new Map();
  const results: Map<string, number> = new Map();
  let max = 0;
  const maxKey: string[] = [];

  textResult.forEach((result) => {
    result.forEach((res) => {
      let count = 1;

      if (results.has(res)) {
        count += results.get(res);
      }

      if (count >= max) {
        max = count;
      }
      results.set(res, count);
    });
  });

  results.forEach((count, candidate) => {
    if (count === max) {
      maxKey.push(candidate);
    }

    const percentage = (results.get(candidate) / textResult.length) * 100;
    const roundedPercentage = (Math.round(percentage * 100) / 100).toFixed(2);
    resultsInPercent.set(candidate, roundedPercentage);
  });

  return { resultsInPercent, maxKey };
};

export { countRankResult, countSelectResult, countTextResult };
