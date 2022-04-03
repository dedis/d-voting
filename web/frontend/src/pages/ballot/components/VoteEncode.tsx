import { Buffer } from 'buffer';
import ShortUniqueId from 'short-unique-id';
import { Answers, RANK, SELECT, TEXT } from 'types/configuration';

export function voteEncode(answers: Answers, maxBallotSize: number) {
  var encodedBallot = '';

  answers.SelectAnswers.forEach((select) => {
    encodedBallot += SELECT + ':' + select.ID + ':';
    select.Answers.forEach((answer) => (encodedBallot += answer ? '1,' : '0,'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.RankAnswers.forEach((rank) => {
    encodedBallot += RANK + ':' + rank.ID + ':';
    let position = Array<number>(rank.Answers.length);
    for (let i = 0; i < rank.Answers.length; i++) {
      position[rank.Answers[i]] = i;
    }
    position.forEach((pos) => (encodedBallot += pos + ','));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.TextAnswers.forEach((text) => {
    encodedBallot += TEXT + ':' + text.ID + ':';
    text.Answers.forEach((answer) => (encodedBallot += 'base64("' + answer + '"),'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  encodedBallot += '\n';

  let ballotSize = Buffer.byteLength(encodedBallot);
  // add padding if necessary
  if (ballotSize < maxBallotSize) {
    const padding = new ShortUniqueId({ length: maxBallotSize - ballotSize });
    encodedBallot += padding();
  }

  var ballotChunks = Array<string>();
  // divide the concatenated string into chunks of 29 bytes
  for (let i = 0; i < maxBallotSize; i += 29) {
    ballotChunks.push(encodedBallot.substring(i, i + 29));
  }

  return ballotChunks;
}
