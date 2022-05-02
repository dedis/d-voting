import { AuthContext } from 'index';
import { FC, useContext } from 'react';
import { STATUS } from 'types/election';
import { ROLE } from 'types/userRole';

type StatusTimelineProps = {
  status: STATUS;
};

const CanceledStep = { name: 'Canceled', state: 'upcoming', status: STATUS.Canceled };

const StatusTimeline: FC<StatusTimelineProps> = ({ status }) => {
  const authCtx = useContext(AuthContext);

  const completeSteps = [
    { name: 'Initial', state: 'complete', status: STATUS.Initial },
    { name: 'Initialized', state: 'current', status: STATUS.InitializedNodes },
    { name: 'On Going Setup', state: 'upcoming', status: STATUS.Setup },
    { name: 'Setup', state: 'upcoming', status: STATUS.Setup },
    { name: 'Open', state: 'upcoming', status: STATUS.Open },
    { name: 'Closed', state: 'upcoming', status: STATUS.Closed },
    { name: 'On Going Shuffle', state: 'upcoming', status: STATUS.OnGoingShuffle },
    { name: 'Shuffled Ballots', state: 'upcoming', status: STATUS.ShuffledBallots },
    { name: 'On Going Decryption', state: 'upcoming', status: STATUS.OnGoingDecryption },
    { name: 'Decrypted Ballots', state: 'upcoming', status: STATUS.DecryptedBallots },
    { name: 'Result Available', state: 'upcoming', status: STATUS.ResultAvailable },
  ];

  const simpleSteps = [
    { name: 'Initial', state: 'complete', status: STATUS.Initial },
    { name: 'Open', state: 'current', status: STATUS.Open },
    { name: 'Closed', state: 'upcoming', status: STATUS.Closed },
    { name: 'Shuffled Ballots', state: 'upcoming', status: STATUS.ShuffledBallots },
    { name: 'Decrypted Ballots', state: 'upcoming', status: STATUS.DecryptedBallots },
    { name: 'Result Available', state: 'upcoming', status: STATUS.ResultAvailable },
  ];

  const steps =
    authCtx.role === ROLE.Admin || authCtx.role === ROLE.Operator ? completeSteps : simpleSteps;

  // If the status is Canceled we need to add the Canceled step to the steps
  // array at the correct position in the workflow (after the Open step)
  if (status === STATUS.Canceled) {
    steps.splice(
      steps.findIndex((step) => step.status === STATUS.Closed),
      0,
      CanceledStep
    );
  }

  const currentStep = steps.findIndex((step) => step.status === status);

  const DisplayStatus = ({ state, name }) => {
    switch (state) {
      case 'complete':
        return (
          <>
            <div className="group pl-4 py-2 flex flex-col border-l-4 border-indigo-600 hover:border-indigo-800 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4">
              <span className="text-xs text-indigo-600 font-semibold tracking-wide uppercase group-hover:text-indigo-800">
                {name}
              </span>
            </div>
          </>
        );
      case 'current':
        return (
          <>
            <div
              className=" animate-pulse pl-4 py-2 flex flex-col border-l-4 border-indigo-600 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4"
              aria-current="step">
              <span className="animate-pulse text-xs text-indigo-600 font-semibold tracking-wide uppercase">
                {name}
              </span>
            </div>
          </>
        );
      default:
        return (
          <>
            <div className="group pl-4 py-2 flex flex-col border-l-4 border-gray-200 hover:border-gray-300 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4">
              <span className="text-xs text-gray-500 font-semibold tracking-wide uppercase group-hover:text-gray-700">
                {name}
              </span>
            </div>
          </>
        );
    }
  };

  return (
    <ol className="space-y-1 md:flex md:space-y-0 md:space-x-2 ">
      {steps.map((step, index) => {
        if (index < currentStep) {
          return (
            <li key={step.name} className="md:flex-1">
              <DisplayStatus state={'complete'} name={step.name} />
            </li>
          );
        }
        if (index === currentStep) {
          return (
            <li key={step.name} className="md:flex-1">
              <DisplayStatus state={'current'} name={step.name} />
            </li>
          );
        }
        return (
          <li key={step.name} className="md:flex-1">
            <DisplayStatus state={'coming'} name={step.name} />
          </li>
        );
      })}
    </ol>
  );
};

export default StatusTimeline;
