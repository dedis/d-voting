import { MinusIcon } from '@heroicons/react/outline';
import PropTypes from 'prop-types';
import { FC } from 'react';
import { useTranslation } from 'react-i18next';
import { NodeStatus } from 'types/node';

type DKGStatusProps = {
  status: NodeStatus;
};

const DKGStatus: FC<DKGStatusProps> = ({ status }) => {
  const { t } = useTranslation();

  const DisplayStatus = () => {
    switch (status) {
      case NodeStatus.Unreachable:
        return (
          <div className="flex items-center">
            <div>
              <MinusIcon className="ml-2 mr-2 h-5 w-5 text-gray-600" aria-hidden="true" />
            </div>
          </div>
        );
      case NodeStatus.NotInitialized:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-gray-500 rounded-full mr-2"></div>
            <div className="max-w-[16vw] truncate">{t('uninitialized')}</div>
          </div>
        );
      case NodeStatus.Initialized:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('initialized')}</div>
          </div>
        );
      case NodeStatus.Setup:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-green-700 rounded-full mr-2"></div>
            <div>{t('statusSetup')}</div>
          </div>
        );
      case NodeStatus.Failed:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-red-500 rounded-full mr-2"></div>
            <div>{t('failed')}</div>
          </div>
        );
      case NodeStatus.Dealing:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-blue-500 rounded-full mr-2"></div>
            <div>{t('dealing')}</div>
          </div>
        );
      case NodeStatus.Responding:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-blue-500 rounded-full mr-2"></div>
            <div>{t('responding')}</div>
          </div>
        );
      case NodeStatus.Certifying:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-blue-500 rounded-full mr-2"></div>
            <div>{t('certifying')}</div>
          </div>
        );
      case NodeStatus.Certified:
        return (
          <div className="flex items-center">
            <div className="block h-4 w-4 bg-green-500 rounded-full mr-2"></div>
            <div>{t('certified')}</div>
          </div>
        );
      default:
        return null;
    }
  };

  return <div className="inline-block align-left">{DisplayStatus()}</div>;
};

DKGStatus.propTypes = {
  status: PropTypes.number,
};

export default DKGStatus;
