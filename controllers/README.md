# Controller Design of Chaos Mesh

This document describes controllers common specification in Chaos Mesh. Although no "standard" should be considered as
absolute requirements (and the real world is full of trade-off and corner case), they should be carefully considered when
you are trying to add a new controller.

## One controller per field

One field should only be "controlled" by at most one controller. In this chapter, multiple reasons will be listed for
this design:

### Avoid the hidden bugs

Multiple controller modifying single object could lead to conflict situation (which is more like a global optimistic lock).
The common way to solve conflict is to adapt the modification and retry. However, if multiple controllers want to modify
a single field, how could they merge the conflict? What's more, it always leads to a hidden bug under the logic. Here is an
example:

If you want to split "pause" and "duration" (the former common chaos) into two standalone controller, let's try to describe
the logic of them:

For the "pause" controller, when the annotation is added, the chaos should enter "not injected" mode, and when the annotation
is removed, the chaos should enter "injected" mode.

For the "duration" controller, when the time exceed the duration, the chaos should enter "not injected" mode.

Though these logics seem to be intuitive, there is a bug under the conflict "mode" (or the `desiredPhase` in current code).
What will happen if the user remove the annotation after the duration exceed? The chaos will enter "injected" and then turn
into "not injected" mode (with the help of "duration" controller), which is dirty and confusing.

If we obey "One field per controller" rule, then they should be combined into one controller and can never be split.

### Handle the conflict in an easier way

After retry the conflict error, we don't need to rerun the whole controller logic (as there may be some side effect in the
controller). Instead, we could save the single field, and set the corresponding field after getting the new object. Which
will give us more confidence on the retry attempting.

## Controller should work standalone

The behavior of every controller should be defined carefully, and they should be able to work without other controller.
The behavior of the controller should also be simple and easy to understand. Try to conclude the action/logic of the controller
in one hundred words, if you failed, please reconsider whether it should be "one" controller, but not two or more (or even
split a new CustomResource).

## Controller should be well documented

Every controller should be described with a "little"/"short" document.
