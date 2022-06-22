import React, { FC } from 'react';
import { LightElectionInfo } from 'types/election';
import { Link } from 'react-router-dom';
import ElectionStatus from './ElectionStatus';
import QuickAction from './QuickAction';

type ElectionRowProps = {
  election: LightElectionInfo;
};

const ElectionRow: FC<ElectionRowProps> = ({ election }) => {
  return (
    <tr className="bg-white border-b hover:bg-gray-50 ">
      <td className="px-1.5 sm:px-6 py-4 font-medium text-gray-900 whitespace-nowrap truncate">
        <Link
          className="text-gray-700 hover:text-indigo-500"
          to={`/elections/${election.ElectionID}`}>
          <div className="max-w-[20vw] truncate">{election.Title}</div>
        </Link>
      </td>
      <td className="px-1.5 sm:px-6 py-4">{<ElectionStatus status={election.Status} />}</td>
      <td className="px-1.5 sm:px-6 py-4 text-right">
        <QuickAction status={election.Status} electionID={election.ElectionID} />
      </td>
    </tr>
  );
};

export default ElectionRow;
