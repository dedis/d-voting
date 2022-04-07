import { Buffer } from 'buffer';
import ShortUniqueId from 'short-unique-id';
import { Answers, RANK, SELECT, TEXT } from 'types/configuration';

export function voteEncode(answers: Answers, maxBallotSize: number, chunksPerBallot: number) {
  var encodedBallot = '';

  answers.SelectAnswers.forEach((selectAnswer, id) => {
    encodedBallot += SELECT + ':' + id + ':';
    selectAnswer.forEach((answer) => (encodedBallot += answer ? '1,' : '0,'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.RankAnswers.forEach((rankAnswer, id) => {
    encodedBallot += RANK + ':' + id + ':';
    let position = Array<number>(rankAnswer.length);
    for (let i = 0; i < rankAnswer.length; i++) {
      position[rankAnswer[i]] = i;
    }
    position.forEach((pos) => (encodedBallot += pos + ','));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.TextAnswers.forEach((textAnswer, id) => {
    encodedBallot += TEXT + ':' + id + ':';
    textAnswer.forEach((answer) => (encodedBallot += 'base64("' + answer + '"),'));
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
  let chunkSize = maxBallotSize / chunksPerBallot;
  // divide the concatenated string into chunks of 29 bytes
  for (let i = 0; i < maxBallotSize; i += chunkSize) {
    ballotChunks.push(encodedBallot.substring(i, i + chunkSize));
  }

  return ballotChunks;
}
