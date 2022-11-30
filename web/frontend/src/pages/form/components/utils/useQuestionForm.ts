import { useState } from 'react';
import { RankQuestion, SelectQuestion, TextQuestion } from 'types/configuration';

// form hook that handles the form state for all types of questions
const useQuestionForm = (initState: RankQuestion | SelectQuestion | TextQuestion) => {
  const [state, setState] = useState<RankQuestion | SelectQuestion | TextQuestion>(initState);
  const { MinN, Choices} = state;

  // depending on the type of the Exception in the question, the form state is
  // updated accordingly
  const handleChange =
    (Exception?: string, optionnalValues?: number) => (e?: React.ChangeEvent<HTMLInputElement>) => {
      const { value, type, name } = e.target;
      const newChoices = Choices;
      switch (Exception) {
        case 'RankMinMax':
          setState({ ...state, MinN: Number(value), MaxN: Number(value) });
          break;
        case 'addChoiceRank':
          console.log('c')
          newChoices.set('en', [...newChoices.get('en'),''])
          newChoices.set('fr', [...newChoices.get('fr'),''])
          newChoices.set('de', [...newChoices.get('de'),''])
          setState({
            ...state,
            Choices: {...newChoices},
            MaxN: Choices.get('en').length + 1,
            MinN: Choices.get('en').length + 1,
          });
          console.log([...Choices.entries()])
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

          setState({
            ...state,
            Choices: new Map([['en' , filteredChoices], ['fr' , filteredChoicesFr], ['de' , filteredChoicesDe]]),
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
    switch (lang){
        case 'en':
            console.log('a')
            Choices.set('en', [...Choices.get('en'),''])
            setState({...state,Choices : Choices,MaxN: Choices.get('en').length + 1})
            break;
        case 'fr':
            Choices.set('fr', [...Choices.get('fr'),''])
            setState({...state,Choices :Choices,MaxN: Choices.get('fr').length + 1})
            break;
        case 'de':
            Choices.set('de', [...Choices.get('de'),''])
            setState({...state,Choices:Choices,MaxN: Choices.get('de').length + 1});
            break;
        default :   
            Choices.set('en', [...Choices.get('en'),'']) 
            setState({...state,Choices : Choices,MaxN: Choices.get('en').length + 1})
            
    }
  };

  // remove a choice from the choices array
  const deleteChoice = (index: number) => {
    if (Choices.get('en').length > MinN) {
      const filteredChoices = Choices.get('en').filter((item: string, idx: number) => idx !== index);
      setState({
        ...state,
        Choices: Choices.set('en',filteredChoices),
        MaxN: Math.max(filteredChoices.length + 1,Choices.get('fr').length + 1, Choices.get('de').length + 1),
      });
    }
    if (Choices.get('fr').length > MinN) {
        const filteredChoicesFr = Choices.get('fr').filter((item: string, idx: number) => idx !== index);   
        setState({
            ...state,
            Choices: Choices.set('fr',filteredChoicesFr),
            MaxN: Math.max(Choices.get('en').length + 1,filteredChoicesFr.length + 1, Choices.get('en').length + 1),
          });
    } 
    if (Choices.get('de').length > MinN) {
        const filteredChoicesDe = Choices.get('en').filter((item: string, idx: number) => idx !== index);    
        setState({
           ...state,
           Choices: Choices.set('de',filteredChoicesDe),
           MaxN: Math.max(Choices.get('en').length + 1,Choices.get('en').length + 1, filteredChoicesDe.length + 1),
         });
    }};

  // update the choice at the given index
  const updateChoice = (index: number,lang: string) => (e) => {
    e.persist();
    const newChoices = Choices;
    switch (lang){
        
        case 'en' :
            console.log('b')
            const choice = newChoices.get('en').map((item: string, idx: number) => {
                if (idx === index) {
                return e.target.value;
                }
                return item;
            })
            newChoices.set('en', choice)
            console.log([...newChoices.entries()])
            console.log(newChoices.get('en'))
            break
        case 'fr' :
            newChoices.set('fr',newChoices.get('fr').map((item: string, idx: number) => {
                  if (idx === index) {
                    return e.target.value;
                  }
                  return item;
                }))
                break
        case 'de' : 
            newChoices.set('de',newChoices.get('de').map((item: string, idx: number) => {
                if (idx === index) {
                    return e.target.value;
                }
                return item
                }))
         
            break
        default:
          newChoices.set('en', newChoices.get('en').map((item: string, idx: number) => {
            if (idx === index) {
            return e.target.value;
            }
            return item;
        }))
        
        
    };
    setState({...state,Choices :  {...newChoices }})  
    console.log([...Choices.entries()])
  };

  return { state, handleChange, addChoice, deleteChoice, updateChoice };
};

export default useQuestionForm;
