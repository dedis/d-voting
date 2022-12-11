import { useEffect, useState } from 'react';
import { RankQuestion, SelectQuestion, TextQuestion } from 'types/configuration';

// form hook that handles the form state for all types of questions
const useQuestionForm = (initState: RankQuestion | SelectQuestion | TextQuestion) => {
  const [state, setState] = useState<RankQuestion | SelectQuestion | TextQuestion>(initState);
  const { MinN, Choices } = state;

  // depending on the type of the Exception in the question, the form state is
  // updated accordingly
  const handleChange =
    (Exception?: string, optionnalValues?: number) => (e?: React.ChangeEvent<HTMLInputElement>) => {
      const { value, type, name } = e.target;
      const obj = Object.fromEntries(Choices)
      const newChoices = new Map(Object.entries(obj))
      newChoices.set('en', [...newChoices.get('en'),''])
      newChoices.set('fr', [...newChoices.get('fr'),''])
      newChoices.set('de', [...newChoices.get('de'),''])
      switch (Exception) {
        case 'RankMinMax':
          setState({ ...state, MinN: Number(value), MaxN: Number(value) });
          break;
        case 'addChoiceRank':
          setState({
            ...state,
            Choices: newChoices,
            MaxN: Math.max(Choices.get('en').length + 1, Choices.get('fr').length + 1, Choices.get('de').length + 1),
            MinN: Math.min(Choices.get('en').length + 1, Choices.get('fr').length + 1, Choices.get('de').length + 1),
          });
          break;
        case 'deleteChoiceRank':
          const filteredChoices = Choices.get('en').filter(
            (item: string, idx: number) => idx !== optionnalValues
          );
          const filteredChoicesFr = Choices.get('fr').filter(
            (item: string, idx: number) => idx !== optionnalValues
          );
          const filteredChoicesDe = Choices.get('de').filter(
            (item: string, idx: number) => idx !== optionnalValues
          );
          const newState = {'en': filteredChoices, 'fr': filteredChoicesFr, 'de': filteredChoicesDe}
          setState({
            ...state,
            Choices: new Map(Object.entries(newState)),
            MaxN: filteredChoices.length,
            MinN: filteredChoices.length,
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

  // updates the choices array when the user adds a new choice
  const addChoice = (lang) => {
    const obj = Object.fromEntries(Choices)
    const newChoices = new Map(Object.entries(obj))
    switch (lang) {
      case 'en':
        setState({ ...state, Choices:newChoices.set('en', [...newChoices.get('en'),'']) , MaxN: Choices.get('en').length + 1 });
        break;
      case 'fr':
        setState({ ...state, Choices :newChoices.set('fr', [...newChoices.get('fr'),'']) , MaxN: Choices.get('fr').length + 1 });
        break;
      case 'de':
        setState({ ...state, Choices :newChoices.set('de', [...newChoices.get('de'),'']) , MaxN: Choices.get('de').length + 1 });
        break;
      default:
        setState({ ...state, Choices:newChoices.set('en', [...newChoices.get('en'),'']) , MaxN: Choices.get('en').length + 1 });
    }
  };

  // remove a choice from the choices array
  const deleteChoice = (index: number) => {
    if (Choices.get('en').length > MinN) {
      const filteredChoices = Choices.get('en').filter((item: string, idx: number) => idx !== index);
      setState({
        ...state,
        Choices: Choices.set('en', filteredChoices),
        MaxN: Math.max(filteredChoices.length + 1, Choices.get('fr').length + 1, Choices.get('de').length + 1),
      });
    }
    if (Choices.get('fr').length > MinN) {
      const filteredChoicesFr = Choices.get('fr').filter((item: string, idx: number) => idx !== index);
      setState({
        ...state,
        Choices: Choices.set('fr', filteredChoicesFr),
        MaxN: Math.max(Choices.get('en').length + 1, filteredChoicesFr.length + 1, Choices.get('de').length + 1),
      });
    }
    if (Choices.get('de').length > MinN) {
      const filteredChoicesDe = Choices.get('de').filter((item: string, idx: number) => idx !== index);
      setState({
        ...state,
        Choices: Choices.set('de', filteredChoicesDe) ,
        MaxN: Math.max(Choices.get('en').length + 1, Choices.get('de').length + 1, filteredChoicesDe.length + 1),
      });
    }
  };
  useEffect (() => {
    console.log('Choices',[...Choices.entries()])
  },[Choices]) 
  // update the choice at the given index
  const updateChoice = (index: number, lang: string) => (e) => {
    e.persist(); 
    const obj = Object.fromEntries(Choices)
    const newChoices = new Map(Object.entries(obj)) 
    switch (lang) {
      case 'en':
        const choice = newChoices.get('en').map((item: string, idx: number) => {
          console.log(e.target.value)
          console.log(index)
          if (idx === index) {
            return e.target.value;
          }
          return item;
        })
        newChoices.set('en',choice)
        setState({
          ...state,
          Choices: newChoices,
        });
        console.log('new', [...newChoices.entries()])
        break;
      case 'fr':
        setState({
          ...state,
          Choices: newChoices.set('fr',newChoices.get('fr').map((item: string, idx: number) => {
            if (idx === index) {
              return e.target.value;
            }
            return item;
          })),
        });
        break;
      case 'de':
        setState({
          ...state,
          Choices: newChoices.set('de',newChoices.get('de').map((item: string, idx: number) => {
            if (idx === index) {
              return e.target.value;
            }
            return item;
          })),
        });
        break;
      default:
        setState({
          ...state,
          Choices: newChoices.set('en',newChoices.get('en').map((item: string, idx: number) => {
            if (idx === index) {
              return e.target.value;
            }
            return item;
          })),
        });
    }
    
  };
  //console.log('end', [...Choices.entries()])
  console.log(state)
  return { state, handleChange, addChoice, deleteChoice, updateChoice };
};

export default useQuestionForm;
