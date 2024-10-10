import { RankQuestion } from 'types/configuration';

// Sum the position for each candidate such that a low score is better
// (e.g if choice 1 ranked first and then fourth then it will have a
// score of (1-1 + 4-1) = 3) and returns the counts and the candidate(s)
// with the lowest score
const countRankResult = (rankResult: number[][], rank: RankQuestion) => {
  const resultsInPercent: string[] = [];
  const minIndices: number[] = [];
  // the maximum score achievable is (number of choices - 1) * number of ballots

  let min = (rank.ChoicesMap.ChoicesMap.get('en').length - 1) * rankResult.length;

  const results = rankResult.reduce((a, b) => {
    return a.map((value, index) => {
      return value + b[index];
    });
  });

  // Total number of "points" attributed
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

// Count the number of vote for a candidate and returns the counts as a
// percentage of the total number of votes and which candidate(s) in the
// select.Choices has the most votes
const countSelectResult = (selectResult: number[][]) => {
  const results: [string, number][] = [];

  selectResult
    .reduce(
      (tally, currBallot) => tally.map((currCount, index) => currCount + currBallot[index]),
      new Array(selectResult[0].length).fill(0)
    )
    .forEach((totalCount) => {
      results.push([
        (Math.round((totalCount / selectResult.length) * 100 * 100) / 100).toFixed(2).toString(),
        totalCount,
      ]);
    });
  return results;
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
