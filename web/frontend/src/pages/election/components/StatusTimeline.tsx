import { AuthContext } from 'index';
import { FC, useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { OngoingAction, Status } from 'types/election';
import { UserRole } from 'types/userRole';

type StatusTimelineProps = {
  status: Status;
  ongoingAction: OngoingAction;
};

const CanceledStep = { name: 'canceled', ongoing: 'canceling', status: Status.Canceled };

const StatusTimeline: FC<StatusTimelineProps> = ({ status, ongoingAction }) => {
  const authCtx = useContext(AuthContext);
  const { t } = useTranslation();

  const completeSteps = [
    { name: 'statusInitial', ongoing: 'statusInitial', status: Status.Initial },
    { name: 'statusInitializedNodes', ongoing: 'initializing', status: Status.Initialized },
    { name: 'statusSetup', ongoing: 'settingUp', status: Status.Setup },
    { name: 'statusOpen', ongoing: 'opening', status: Status.Open },
    { name: 'statusClose', ongoing: 'closing', status: Status.Closed },
    { name: 'statusShuffle', ongoing: 'shuffling', status: Status.ShuffledBallots },
    { name: 'statusDecrypted', ongoing: 'decrypting', status: Status.PubSharesSubmitted },
    { name: 'statusResultAvailable', ongoing: 'combining', status: Status.ResultAvailable },
  ];

  const simpleSteps = [
    { name: 'statusInitial', ongoing: 'statusInitial', status: Status.Initial },
    { name: 'statusOpen', ongoing: 'opening', status: Status.Open },
    { name: 'statusClose', ongoing: 'closing', status: Status.Closed },
    { name: 'statusShuffle', ongoing: 'shuffling', status: Status.ShuffledBallots },
    { name: 'statusDecrypted', ongoing: 'decrypting', status: Status.PubSharesSubmitted },
    { name: 'statusResultAvailable', ongoing: 'combining', status: Status.ResultAvailable },
  ];

  const steps =
    authCtx.role === UserRole.Admin || authCtx.role === UserRole.Operator
      ? completeSteps
      : simpleSteps;

  // If the status is Canceled we need to add the Canceled step to the steps
  // array at the correct position in the workflow (before the Closed step)
  if (status === Status.Canceled) {
    steps.splice(
      steps.findIndex((step) => step.status === Status.Closed),
      0,
      CanceledStep
    );
  }

  // Find the current step in the steps array (the status)
  const currentStep = steps.findIndex((step) => step.status === status);

  const DisplayStatus = ({ state, step, index }) => {
    switch (state) {
      case 'complete':
        return (
          <div className="group pl-4 py-2 flex flex-col border-l-4 border-indigo-600 hover:border-indigo-800 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4">
            <span className="text-xs text-indigo-600 font-semibold tracking-wide uppercase group-hover:text-indigo-800">
              {t(step.name)}
            </span>
          </div>
        );
      case 'current':
        return (
          <div
            className="pl-4 py-2 flex flex-col border-l-4 border-indigo-600 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4"
            aria-current="step">
            <span className="text-xs text-indigo-600 font-semibold tracking-wide uppercase">
              {t(step.name)}
            </span>
          </div>
        );
      default:
        if (ongoingAction === index) {
          return (
            <div
              className="animate-pulse pl-4 py-2 flex flex-col border-l-4 border-indigo-400 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4"
              aria-current="step">
              <span className="text-xs text-indigo-400 font-semibold tracking-wide uppercase">
                {t(step.ongoing)}
              </span>
            </div>
          );
        }
        return (
          <div className="group pl-4 py-2 flex flex-col border-l-4 border-gray-200 hover:border-gray-300 md:pl-0 md:pt-4 md:pb-0 md:border-l-0 md:border-t-4">
            <span className="text-xs text-gray-500 font-semibold tracking-wide uppercase group-hover:text-gray-700">
              {t(step.name)}
            </span>
          </div>
        );
    }
  };

  return (
    <div>
      <ol className="space-y-2 md:flex md:space-y-0 md:space-x-2 ">
        {steps.map((step, index) => {
          if (index < currentStep) {
            return (
              <li key={step.name} className="md:flex-1">
                <DisplayStatus state={'complete'} step={step} index={index} />
              </li>
            );
          }
          if (index === currentStep) {
            return (
              <li key={step.name} className="md:flex-1">
                <DisplayStatus state={'current'} step={step} index={index} />
              </li>
            );
          }
          return (
            <li key={step.name} className="md:flex-1">
              <DisplayStatus state={'coming'} step={step} index={index} />
            </li>
          );
        })}
      </ol>
    </div>
  );
};

export default StatusTimeline;
