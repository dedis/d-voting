import { Answers } from 'components/utils/useConfiguration';
import { Buffer } from 'buffer';
import ShortUniqueId from 'short-unique-id';

export function voteEncode(answers: Answers, ballotSize: number) {
  var encodedBallot = '';

  answers.SelectAnswers.forEach((select) => {
    encodedBallot += 'select:' + select.ID + ':';
    select.Answers.forEach((answer) => (encodedBallot += answer ? '1,' : '0,'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.RankAnswers.forEach((rank) => {
    encodedBallot += 'rank:' + rank.ID + ':';
    rank.Answers.forEach((answer) => (encodedBallot += answer + ',')); // is this the right format ?
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  answers.TextAnswers.forEach((text) => {
    encodedBallot += 'text:' + text.ID + ':';
    text.Answers.forEach((answer) => (encodedBallot += 'base64("' + answer + '"),'));
    encodedBallot = encodedBallot.slice(0, -1);
    encodedBallot += '\n';
  });

  encodedBallot += '\n';

  let byteSize = Buffer.byteLength(encodedBallot);
  if (byteSize < ballotSize) {
    const padding = new ShortUniqueId({ length: ballotSize - byteSize });
    encodedBallot += padding();
  }
  console.log('encodedBallot:' + encodedBallot);

  var ballotChunks = Array<string>();
  for (let i = 0; i < ballotSize; i += 29) {
    ballotChunks.push(encodedBallot.substring(i, i + 29));
  }

  return ballotChunks;
}
