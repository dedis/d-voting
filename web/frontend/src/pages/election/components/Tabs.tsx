function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(' ');
}

const tabs = [
  { name: 'electionForm', title: 'Election' },
  { name: 'previewForm', title: 'Preview' },
];

const Tabs = ({ currentTab, setCurrentTab }) => {
  return (
    <div className="mx-4 mb-6">
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex" aria-label="Tabs">
          {tabs.map(({ name, title }) => (
            <button
              key={name}
              onClick={() => {
                setCurrentTab(name);
              }}
              className={classNames(
                name === currentTab
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
                'w-1/4 py-4 px-1 text-center border-b-2 font-medium text-sm'
              )}
              aria-current={name === currentTab ? 'page' : undefined}>
              {title}
            </button>
          ))}
        </nav>
      </div>
    </div>
  );
};

export default Tabs;
