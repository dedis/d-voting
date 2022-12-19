import { useEffect, useState } from 'react';
import { RankQuestion, SelectQuestion, TextQuestion } from 'types/configuration';
import { choicesMapToChoices } from '../../../../types/getObjectType';

// form hook that handles the form state for all types of questions
const useQuestionForm = (initState: RankQuestion | SelectQuestion | TextQuestion) => {
  const [state, setState] = useState<RankQuestion | SelectQuestion | TextQuestion>(initState);
  const { MinN, ChoicesMap } = state;

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
          const filteredChoicesMap = ChoicesMap.get('en').filter(
            (item: string, idx: number) => idx !== optionnalValues
          );
          const filteredChoicesMapFr = ChoicesMap.get('fr').filter(
            (item: string, idx: number) => idx !== optionnalValues
          );
          const filteredChoicesMapDe = ChoicesMap.get('de').filter(
            (item: string, idx: number) => idx !== optionnalValues
          );
          const newState = { en: filteredChoicesMap, fr: filteredChoicesMapFr, de: filteredChoicesMapDe };
          setState({
            ...state,
            ChoicesMap: new Map(Object.entries(newState)),
            MaxN: filteredChoicesMap.length,
            MinN: filteredChoicesMap.length,
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

  // updates the ChoicesMap array when the user adds a new choice
  const addChoice = (lang) => {
    const obj = Object.fromEntries(ChoicesMap);
    const newChoicesMap = new Map(Object.entries(obj));
    switch (lang) {
      case 'en':
        setState({
          ...state,
          ChoicesMap: newChoicesMap.set('en', [...newChoicesMap.get('en'), '']),
          MaxN: ChoicesMap.get('en').length + 1,
        });
        break;
      case 'fr':
        setState({
          ...state,
          ChoicesMap: newChoicesMap.set('fr', [...newChoicesMap.get('fr'), '']),
          MaxN: ChoicesMap.get('fr').length + 1,
        });
        break;
      case 'de':
        setState({
          ...state,
          ChoicesMap: newChoicesMap.set('de', [...newChoicesMap.get('de'), '']),
          MaxN: ChoicesMap.get('de').length + 1,
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

  // remove a choice from the ChoicesMap array
  const deleteChoice = (index: number) => {
    if (ChoicesMap.get('en').length > MinN) {
      const filteredChoicesMap = ChoicesMap.get('en').filter(
        (item: string, idx: number) => idx !== index
      );
      setState({
        ...state,
        ChoicesMap: ChoicesMap.set('en', filteredChoicesMap),
        MaxN: Math.max(
          filteredChoicesMap.length + 1,
          ChoicesMap.get('fr').length + 1,
          ChoicesMap.get('de').length + 1
        ),
      });
    }
    if (ChoicesMap.get('fr').length > MinN) {
      const filteredChoicesMapFr = ChoicesMap.get('fr').filter(
        (item: string, idx: number) => idx !== index
      );
      setState({
        ...state,
        ChoicesMap: ChoicesMap.set('fr', filteredChoicesMapFr),
        MaxN: Math.max(
          ChoicesMap.get('en').length + 1,
          filteredChoicesMapFr.length + 1,
          ChoicesMap.get('de').length + 1
        ),
      });
    }
    if (ChoicesMap.get('de').length > MinN) {
      const filteredChoicesMapDe = ChoicesMap.get('de').filter(
        (item: string, idx: number) => idx !== index
      );
      setState({
        ...state,
        ChoicesMap: ChoicesMap.set('de', filteredChoicesMapDe),
        MaxN: Math.max(
          ChoicesMap.get('en').length + 1,
          ChoicesMap.get('de').length + 1,
          filteredChoicesMapDe.length + 1
        ),
      });
    }
  };
  // update the choice at the given index
  const updateChoice = (index: number, lang: string) => (e) => {
    e.persist();
    const obj = Object.fromEntries(ChoicesMap);
    switch (lang) {
      case 'en':
      case 'fr':
      case 'de':
        const newChoicesMap = new Map(Object.entries({
          ...obj,
          [lang]: obj[lang].map((item: string, idx: number) =>
          (idx === index ? e.target.value : item)),
        }));
        setState({
          ...state,
          ChoicesMap: newChoicesMap,
          Choices: choicesMapToChoices(newChoicesMap),
        });
        break;
      default:
        console.error('WROOONG')
    }

  };
  console.log('state',state);
  return { state, handleChange, addChoice, deleteChoice, updateChoice };
};

export default useQuestionForm;
