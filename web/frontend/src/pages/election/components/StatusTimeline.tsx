import { AuthContext } from 'index';
import { FC, useContext } from 'react';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

type StatusTimelineProps = {
  status: STATUS;
};

const completeSteps = [
  { name: 'Initial', status: 'complete' },
  { name: 'Initialized', status: 'current' },
  { name: 'On Going Setup', status: 'upcoming' },
  { name: 'Setup', status: 'upcoming' },
  { name: 'Open', status: 'upcoming' },
  // { name: 'Canceled', status: 'upcoming' },
  { name: 'Closed', status: 'upcoming' },
  { name: 'On Going Shuffle', status: 'upcoming' },
  { name: 'Shuffled Ballots', status: 'upcoming' },
  { name: 'On Going Decryption', status: 'upcoming' },
  { name: 'Decrypted Ballots', status: 'upcoming' },
  { name: 'Result Available', status: 'upcoming' },
];

const simpleSteps = [
  { name: 'Initial', status: 'complete' },
  { name: 'Open', status: 'current' },
  // { name: 'Canceled', status: 'upcoming' },
  { name: 'Closed', status: 'upcoming' },
  { name: 'Shuffled Ballots', status: 'upcoming' },
  { name: 'Decrypted Ballots', status: 'upcoming' },
  { name: 'Result Available', status: 'upcoming' },
];

const StatusTimeline: FC<StatusTimelineProps> = ({ status }) => {
  const authCtx = useContext(AuthContext);
  const steps =
    authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator ? completeSteps : simpleSteps;

  return (
    <>
      <ol className="space-y-1 md:flex md:space-y-0 md:space-x-2 ">
        {steps.map((step) => (
          <li key={step.name} className="md:flex-1">
            {step.status === 'complete' ? (
              <div className="group pl-4 py-2 flex flex-col border-l-4 border-indigo-600 hover:border-indigo-800 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4">
                <span className="text-xs text-indigo-600 font-semibold tracking-wide uppercase group-hover:text-indigo-800">
                  {step.name}
                </span>
              </div>
            ) : step.status === 'current' ? (
              <div
                className=" animate-pulse pl-4 py-2 flex flex-col border-l-4 border-indigo-600 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4"
                aria-current="step">
                <span className="animate-pulse text-xs text-indigo-600 font-semibold tracking-wide uppercase">
                  {step.name}
                </span>
              </div>
            ) : (
              <div className="group pl-4 py-2 flex flex-col border-l-4 border-gray-200 hover:border-gray-300 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4">
                <span className="text-xs text-gray-500 font-semibold tracking-wide uppercase group-hover:text-gray-700">
                  {step.name}
                </span>
              </div>
            )}
          </li>
        ))}
      </ol>
    </>
  );
};

export default StatusTimeline;
