<p style="text-align:center"><img width="300px" src="assets/logo.png"/></p>

# D-Voting

## Deployment diagram

The following diagram pictures the d-voting system from a deployment point of
view. It describes the components and their interactions.

<p style="text-align:center"><img src="assets/system.png"/></p>

## Layers

The following layer diagram shows the layers of a node. It illustrates how an
end-user can interact with the system.

<p style="text-align:center"><img src="assets/layers.png"/></p>

## Form flow

The following sequence diagram shows the entire flow of an form.

<p style="text-align:center"><img src="assets/form flow.png"/></p>

<details>
    <summary>source</summary>
title Form flow

actor voter
actor admin
participant smart contract
database global state
database DKGRegistry

== Setup ==

admin->smart contract:OpenForm
smart contract->global state:GetRoster
global state-->smart contract:roster
smart contract->global state:StoreForm(roster, ...)
note over admin:formID can be computed by the admin\nbased on the transaction ID that is unique
admin->DKGRegistry:init(formID)
DKGRegistry->global state:GetForm
global state-->DKGRegistry:form.roster
DKGRegistry->DKGRegistry:dkg = create(roster)
DKGRegistry->DKGRegistry:store(dkg, formID)
admin-->DKGRegistry:setup(formID)

DKGRegistry->DKGRegistry:dkg = get(formID)\npubkey = dkg.setup

== Open ==

admin->smart contract:open(formID)
smart contract->DKGRegistry:GetPubKey(formID)

DKGRegistry-->smart contract:pubkey
smart contract->global state:StoreForm(pubkey, ...)

== Cast ==

voter->global state:GetForm(formID)
global state-->voter:form.pubkey

voter->voter:ballot = encrypt(vote, pubkey)
voter->smart contract:Cast(ballot)

smart contract->global state:StoreForm(ballot, ...)

== Shuffle form ==

admin->smart contract:CloseForm
smart contract->global state:StoreForm(status, ...)

admin->Neff:init
admin->Neff:setup(formID)
Neff->global state:GetForm(formID)
global state-->Neff:form.roster

Neff->smart contract:SubmitShuffle(shuffledBallots)\nuse the transactionID as the random source for the proof
smart contract->global state:StoreForm(shuffledBallots, ...)
smart contract->global state:(if enough shuffling)\nStoreForm(status, ...)

== Terminate ==
admin->DKGRegistry: ComputePubshares()

DKGRegistry->smart contract: SubmitPushares (pubshare)
smart contract->global state: StoreForm(pubshare, ...)

admin->smart contract: CombinePubShares
smart contract->global state: StoreForm(decryptedBallots, ...)

</details>
