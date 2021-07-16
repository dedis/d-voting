
import {React, useContext} from 'react';
import SimpleTable from '../utils/SimpleTable';
import {RESULT_AVAILABLE} from '../utils/StatusNumber';
import {Translations} from '../language/Translations';
import {LanguageContext} from '../language/LanguageContext';

const ResultTable = () => {
    const [context,] = useContext(LanguageContext);
    return (<div>
        <SimpleTable statusToKeep={RESULT_AVAILABLE} pathLink='results' textWhenData={Translations[context].displayResults} textWhenNoData={Translations[context].noResultsAvailable}/>
    </div>)
}

export default ResultTable;

