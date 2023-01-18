import { useState } from 'react';
import { RankQuestion, SelectQuestion, TextQuestion } from 'types/configuration';
import { choicesMapToChoices } from 'types/getObjectType';
import { availableLanguages } from 'language/Configuration';

// form hook that handles the form state for all types of questions
const useQuestionForm = (initState: RankQuestion | SelectQuestion | TextQuestion) => {
  const [state, setState] = useState<RankQuestion | SelectQuestion | TextQuestion>(initState);
  const { MinN, ChoicesMap } = state;
  const lang = availableLanguages;

  // depending on the type of the Exception in the question, the form state is
  // updated accordingly
  const handleChange =
    (Exception?: string, optionnalValues?: number) => (e?: React.ChangeEvent<HTMLInputElement>) => {
      const { value, type, name } = e.target;
      const obj = Object.fromEntries(ChoicesMap);
      const newChoicesMap = new Map(Object.entries(obj));
      newChoicesMap.set('en', [...newChoicesMap.get('en'), '']);
      newChoicesMap.set('fr', [...newChoicesMap.get('fr'), '']);
      newChoicesMap.set('de', [...newChoicesMap.get('de'), '']);
      switch (Exception) {
        case 'RankMinMax':
          setState({ ...state, MinN: Number(value), MaxN: Number(value) });
          break;
        case 'addChoiceRank':
          setState({
            ...state,
            ChoicesMap: newChoicesMap,
            MaxN: Math.max(
              ChoicesMap.get('en').length + 1,
              ChoicesMap.get('fr').length + 1,
              ChoicesMap.get('de').length + 1
            ),
            MinN: Math.min(
              ChoicesMap.get('en').length + 1,
              ChoicesMap.get('fr').length + 1,
              ChoicesMap.get('de').length + 1
            ),
          });
          break;
        case 'deleteChoiceRank':
          lang.forEach((lg) => {
            const filteredChoicesMap = ChoicesMap.get(lg).filter(
              (item: string, idx: number) => idx !== optionnalValues
            );
            ChoicesMap.set(lg, filteredChoicesMap);
          });
          setState({
            ...state,
            ChoicesMap: ChoicesMap,
            MaxN: ChoicesMap.get('en').length,
            MinN: ChoicesMap.get('en').length,
          });
          break;
        case 'TextMaxLength':
          if (Number(value) >= 1) {
            setState({ ...state, MaxLength: Number(value) });
          }
          break;
        default:
          e.persist();
          switch (type) {
            case 'number':
              setState({ ...state, [name]: Number(value) });
              break;
            case 'text':
              setState({ ...state, [name]: value });
              break;
            default:
              break;
          }
          break;
      }
    };

  // updates the ChoicesMap map when the user adds a new choice
  const addChoice = (lg) => {
    const obj = Object.fromEntries(ChoicesMap);
    const newChoicesMap = new Map(Object.entries(obj));
    switch (lg) {
      case lg:
        setState({
          ...state,
          ChoicesMap: newChoicesMap.set(lg, [...newChoicesMap.get(lg), '']),
          MaxN: ChoicesMap.get(lg).length + 1,
        });
        break;
      default:
        setState({
          ...state,
          ChoicesMap: newChoicesMap.set('en', [...newChoicesMap.get('en'), '']),
          MaxN: ChoicesMap.get('en').length + 1,
        });
    }
  };

  // remove a choice from the ChoicesMap map
  const deleteChoice = (index: number) => {
    lang.forEach((lg) => {
      if (ChoicesMap.get(lg).length > MinN) {
        const filteredChoicesMap = ChoicesMap.get(lg).filter(
          (item: string, idx: number) => idx !== index
        );
        ChoicesMap.set(lg, filteredChoicesMap);
      }
    });
    const maxN = Math.max(
      ChoicesMap.get('en').length + 1,
      ChoicesMap.get('fr').length + 1,
      ChoicesMap.get('de').length + 1
    );
    setState({
      ...state,
      ChoicesMap: ChoicesMap,
      MaxN: maxN,
    });
  };

  // update the choice at the given index
  const updateChoice = (index: number, lg: string) => (e) => {
    e.persist();
    const obj = Object.fromEntries(ChoicesMap);
    switch (lg) {
      case lg:
        const newChoicesMap = new Map(
          Object.entries({
            ...obj,
            [lg]: obj[lg].map((item: string, idx: number) =>
              idx === index ? e.target.value : item
            ),
          })
        );
        setState({
          ...state,
          ChoicesMap: newChoicesMap,
          Choices: choicesMapToChoices(newChoicesMap),
        });
        break;

      default:
        const newChoicesMapDefault = new Map(
          Object.entries({
            ...obj,
            ['en']: obj.en.map((item: string, idx: number) =>
              idx === index ? e.target.value : item
            ),
          })
        );
        setState({
          ...state,
          ChoicesMap: newChoicesMapDefault,
          Choices: choicesMapToChoices(newChoicesMapDefault),
        });
    }
  };
  return { state, handleChange, addChoice, deleteChoice, updateChoice };
};

export default useQuestionForm;
