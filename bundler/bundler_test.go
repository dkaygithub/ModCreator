package bundler

import "testing"

func TestUnbundle(t *testing.T) {
	rawlua := `
  -- Bundled by luabundle {"version":"1.6.0"}
  local __bundle_require, __bundle_loaded, __bundle_register, __bundle_modules = (function(superRequire)
  	local loadingPlaceholder = {[{}] = true}

  	local register
  	local modules = {}

  	local require
  	local loaded = {}

  	register = function(name, body)
  		if not modules[name] then
  			modules[name] = body
  		end
  	end

  	require = function(name)
  		local loadedModule = loaded[name]

  		if loadedModule then
  			if loadedModule == loadingPlaceholder then
  				return nil
  			end
  		else
  			if not modules[name] then
  				if not superRequire then
  					local identifier = type(name) == 'string' and '\"' .. name .. '\"' or tostring(name)
  					error('Tried to require ' .. identifier .. ', but no such module has been registered')
  				else
  					return superRequire(name)
  				end
  			end

  			loaded[name] = loadingPlaceholder
  			loadedModule = modules[name](require, loaded, register, modules)
  			loaded[name] = loadedModule
  		end

  		return loadedModule
  	end

  	return require, loaded, register, modules
  end)(nil)
  __bundle_register("__root", function(require, _LOADED, __bundle_register, __bundle_modules)
  require("core/AgendaDeck")
  end)
  __bundle_register("core/AgendaDeck", function(require, _LOADED, __bundle_register, __bundle_modules)
  MIN_VALUE = -99
  MAX_VALUE = 999

  function onload(saved_data)
      light_mode = false
      val = 0

      if saved_data ~= "" then
          local loaded_data = JSON.decode(saved_data)
          light_mode = loaded_data[1]
          val = loaded_data[2]
      end

      createAll()
  end

  function updateSave()
      local data_to_save = {light_mode, val}
      saved_data = JSON.encode(data_to_save)
      self.script_state = saved_data
  end

  function createAll()
      s_color = {0.5, 0.5, 0.5, 95}

      if light_mode then
          f_color = {1,1,1,95}
      else
          f_color = {0,0,0,100}
      end



      self.createButton({
        label=tostring(val),
        click_function="add_subtract",
        function_owner=self,
        position={0,0.05,0},
        height=600,
        width=1000,
        alignment = 3,
        scale={x=1.5, y=1.5, z=1.5},
        font_size=600,
        font_color=f_color,
        color={0,0,0,0}
        })




      if light_mode then
          lightButtonText = "[ Set dark ]"
      else
          lightButtonText = "[ Set light ]"
      end

  end

  function removeAll()
      self.removeInput(0)
      self.removeInput(1)
      self.removeButton(0)
      self.removeButton(1)
      self.removeButton(2)
  end

  function reloadAll()
      removeAll()
      createAll()

      updateSave()
  end

  function swap_fcolor(_obj, _color, alt_click)
      light_mode = not light_mode
      reloadAll()
  end

  function swap_align(_obj, _color, alt_click)
      center_mode = not center_mode
      reloadAll()
  end

  function editName(_obj, _string, value)
      self.setName(value)
      setTooltips()
  end

  function add_subtract(_obj, _color, alt_click)
      mod = alt_click and -1 or 1
      new_value = math.min(math.max(val + mod, MIN_VALUE), MAX_VALUE)
      if val ~= new_value then
          val = new_value
        updateVal()
          updateSave()
      end
  end

  function updateVal()

      self.editButton({
          index = 0,
          label = tostring(val),

          })
  end

  function reset_val()
      val = 0
      updateVal()
      updateSave()
  end

  function setTooltips()
      self.editInput({
          index = 0,
          value = self.getName(),
          tooltip = ttText
          })
      self.editButton({
          index = 0,
          value = tostring(val),
          tooltip = ttText
          })
  end

  function null()
  end

  function keepSample(_obj, _string, value)
      reloadAll()
  end

  end)
  return __bundle_require("__root")
  `
	got, err := Unbundle(rawlua)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	want := `require("core/AgendaDeck")`
	if want != got {
		t.Errorf("want <%s>, got <%s>\n", want, got)
	}
}

func TestUn2(t *testing.T) {
	rawlua := `-- Bundled by luabundle {"version":"1.6.0"}
local __bundle_require, __bundle_loaded, __bundle_register, __bundle_modules = (function(superRequire)
	local loadingPlaceholder = {[{}] = true}

	local register
	local modules = {}

	local require
	local loaded = {}

	register = function(name, body)
		if not modules[name] then
			modules[name] = body
		end
	end

	require = function(name)
		local loadedModule = loaded[name]

		if loadedModule then
			if loadedModule == loadingPlaceholder then
				return nil
			end
		else
			if not modules[name] then
				if not superRequire then
					local identifier = type(name) == 'string' and '\"' .. name .. '\"' or tostring(name)
					error('Tried to require ' .. identifier .. ', but no such module has been registered')
				else
					return superRequire(name)
				end
			end

			loaded[name] = loadingPlaceholder
			loadedModule = modules[name](require, loaded, register, modules)
			loaded[name] = loadedModule
		end

		return loadedModule
	end

	return require, loaded, register, modules
end)(nil)
__bundle_register("__root", function(require, _LOADED, __bundle_register, __bundle_modules)
require("playercards/AllCardsBag")
end)
__bundle_register("playercards/AllCardsBag", function(require, _LOADED, __bundle_register, __bundle_modules)

-- Position to check for weaknesses.  Everything with X and Z less
-- than these values (down and right on the table) will be checked
local WEAKNESS_CHECK_X = 15
local WEAKNESS_CHECK_Z = 37

local cardIdIndex = { }
local classAndLevelIndex = { }
local basicWeaknessList = { }

local indexingDone = false
local allowRemoval = false

function onLoad()
  self.addContextMenuItem("Rebuild Index", startIndexBuild)
  math.randomseed(os.time())
  Wait.frames(startIndexBuild, 30)
end

-- Called by Hotfix bags when they load.  If we are still loading indexes, then
-- the all cards and hotfix bags are being loaded together, and we can ignore
-- this call as the hotfix will be included in the initial indexing.  If it is
-- called once indexing is complete it means the hotfix bag has been added
-- later, and we should rebuild the index to integrate the hotfix bag.
function rebuildIndexForHotfix()
  if (indexingDone) then
    startIndexBuild()
  end
end

-- Resets all current bag indexes
function clearIndexes()
  indexingDone = false
  cardIdIndex = { }
  classAndLevelIndex = { }
  classAndLevelIndex["Guardian-upgrade"] = { }
  classAndLevelIndex["Seeker-upgrade"] = { }
  classAndLevelIndex["Mystic-upgrade"] = { }
  classAndLevelIndex["Survivor-upgrade"] = { }
  classAndLevelIndex["Rogue-upgrade"] = { }
  classAndLevelIndex["Neutral-upgrade"] = { }
  classAndLevelIndex["Guardian-level0"] = { }
  classAndLevelIndex["Seeker-level0"] = { }
  classAndLevelIndex["Mystic-level0"] = { }
  classAndLevelIndex["Survivor-level0"] = { }
  classAndLevelIndex["Rogue-level0"] = { }
  classAndLevelIndex["Neutral-level0"] = { }
  basicWeaknessList = { }
end

-- Clears the bag indexes and starts the coroutine to rebuild the indexes
function startIndexBuild(playerColor)
  clearIndexes()
  startLuaCoroutine(self, "buildIndex")
end

function onObjectLeaveContainer(container, object)
  if (container == self and not allowRemoval) then
    broadcastToAll(
        "Removing cards from the All Player Cards bag may break some functions.  Please replace the card.",
        {0.9, 0.2, 0.2}
    )
  end
end

-- Debug option to suppress the warning when cards are removed from the bag
function setAllowCardRemoval()
  allowRemoval = true
end

-- Create the card indexes by iterating all cards in the bag, parsing their
-- metadata, and creating the keyed lookup tables for the cards.  This is a
-- coroutine which will spread the workload by processing 20 cards before
-- yielding.  Based on the current count of cards this will require
-- approximately 60 frames to complete.
function buildIndex()
  indexingDone = false
  if (self.getData().ContainedObjects == nil) then
    return 1
  end
  for i, cardData in ipairs(self.getData().ContainedObjects) do
    local cardMetadata = JSON.decode(cardData.GMNotes)
    if (cardMetadata ~= nil) then
      addCardToIndex(cardData, cardMetadata)
    end
    if (i % 20 == 0) then
      coroutine.yield(0)
    end
  end
  local hotfixBags = getObjectsWithTag("AllCardsHotfix")
  for _, hotfixBag in ipairs(hotfixBags) do
    if (#hotfixBag.getObjects() > 0) then
      for i, cardData in ipairs(hotfixBag.getData().ContainedObjects) do
        local cardMetadata = JSON.decode(cardData.GMNotes)
        if (cardMetadata ~= nil) then
          addCardToIndex(cardData, cardMetadata)
        end
      end
    end
  end
  buildSupplementalIndexes()
  indexingDone = true
  return 1
end

-- Adds a card to any indexes it should be a part of, based on its metadata.
-- Param cardData: TTS object data for the card
-- Param cardMetadata: SCED metadata for the card
function addCardToIndex(cardData, cardMetadata)
  cardIdIndex[cardMetadata.id] = { data = cardData, metadata = cardMetadata }
  if (cardMetadata.alternate_ids ~= nil) then
    for _, alternateId in ipairs(cardMetadata.alternate_ids) do
      cardIdIndex[alternateId] = { data = cardData, metadata = cardMetadata }
    end
  end
end

function buildSupplementalIndexes()
  for cardId, card in pairs(cardIdIndex) do
    local cardData = card.data
    local cardMetadata = card.metadata
    -- If the ID key and the metadata ID don't match this is a duplicate card created by an
    -- alternate_id, and we should skip it
    if (cardId == cardMetadata.id) then
      -- Add card to the basic weakness list, if appropriate.  Some weaknesses have
      -- multiple copies, and are added multiple times
      if (cardMetadata.weakness and cardMetadata.basicWeaknessCount ~= nil) then
        for i = 1, cardMetadata.basicWeaknessCount do
          table.insert(basicWeaknessList, cardMetadata.id)
      end
    end

  -- Add the card to the appropriate class and level indexes
    local isGuardian = false
    local isSeeker = false
    local isMystic = false
    local isRogue = false
    local isSurvivor = false
    local isNeutral = false
    local upgradeKey
    -- Excludes signature cards (which have no class or level) and alternate
    -- ID entries
    if (cardMetadata.class ~= nil and cardMetadata.level ~= nil) then
        isGuardian = string.match(cardMetadata.class, "Guardian")
        isSeeker = string.match(cardMetadata.class, "Seeker")
        isMystic = string.match(cardMetadata.class, "Mystic")
        isRogue = string.match(cardMetadata.class, "Rogue")
        isSurvivor = string.match(cardMetadata.class, "Survivor")
        isNeutral = string.match(cardMetadata.class, "Neutral")
        if (cardMetadata.level > 0) then
          upgradeKey = "-upgrade"
        else
          upgradeKey = "-level0"
        end
        if (isGuardian) then
          table.insert(classAndLevelIndex["Guardian"..upgradeKey], cardMetadata.id)
        end
        if (isSeeker) then
          table.insert(classAndLevelIndex["Seeker"..upgradeKey], cardMetadata.id)
        end
        if (isMystic) then
          table.insert(classAndLevelIndex["Mystic"..upgradeKey], cardMetadata.id)
        end
        if (isRogue) then
          table.insert(classAndLevelIndex["Rogue"..upgradeKey], cardMetadata.id)
        end
        if (isSurvivor) then
          table.insert(classAndLevelIndex["Survivor"..upgradeKey], cardMetadata.id)
        end
        if (isNeutral) then
          table.insert(classAndLevelIndex["Neutral"..upgradeKey], cardMetadata.id)
        end
      end
    end
  end
  for _, indexTable in pairs(classAndLevelIndex) do
    table.sort(indexTable, cardComparator)
  end
end

-- Comparison function used to sort the class card bag indexes.  Sorts by card
-- level, then name, then subname.
function cardComparator(id1, id2)
  local card1 = cardIdIndex[id1]
  local card2 = cardIdIndex[id2]

  if (card1.metadata.level ~= card2.metadata.level) then
    return card1.metadata.level < card2.metadata.level
  end
  if (card1.data.Nickname ~= card2.data.Nickname) then
    return card1.data.Nickname < card2.data.Nickname
  end
  return card1.data.Description < card2.data.Description
end

function isIndexReady()
  return indexingDone
end

-- Returns a specific card from the bag, based on ArkhamDB ID
-- Params table:
--     id: String ID of the card to retrieve
-- Return: If the indexes are still being constructed, an empty table is
--     returned.  Otherwise, a single table with the following fields
--       cardData: TTS object data, suitable for spawning the card
--       cardMetadata: Table of parsed metadata
function getCardById(params)
  if (not indexingDone) then
    broadcastToAll("Still loading player cards, please try again in a few seconds", {0.9, 0.2, 0.2})
    return { }
  end
  return cardIdIndex[params.id]
end

-- Returns a list of cards from the bag matching a class and level (0 or upgraded)
-- Params table:
--     class: String class to retrieve ("Guardian", "Seeker", etc)
--     isUpgraded: true for upgraded cards (Level 1-5), false for Level 0
-- Return: If the indexes are still being constructed, returns an empty table.
--     Otherwise, a list of tables, each with the following fields
--       cardData: TTS object data, suitable for spawning the card
--       cardMetadata: Table of parsed metadata
function getCardsByClassAndLevel(params)
  if (not indexingDone) then
    broadcastToAll("Still loading player cards, please try again in a few seconds", {0.9, 0.2, 0.2})
    return { }
  end
  local upgradeKey
  if (params.upgraded) then
    upgradeKey = "-upgrade"
  else
    upgradeKey = "-level0"
  end
  return classAndLevelIndex[params.class..upgradeKey];
end

-- Searches the bag for cards which match the given name and returns a list.  Note that this is
-- an O(n) search without index support.  It may be slow.
-- Parameter array must contain these fields to define the search:
--   name String or string fragment to search for names
--   exact Whether the name match should be exact
function getCardsByName(params)
  local name = params.name
  local exact = params.exact
  local results = { }
  -- Track cards (by ID) that we've added to avoid duplicates that may come from alternate IDs
  local addedCards = { }
  for _, cardData in pairs(cardIdIndex) do
    if (not addedCards[cardData.metadata.id]) then
      if (exact and (string.lower(cardData.data.Nickname) == string.lower(name)))
          or (not exact and string.find(string.lower(cardData.data.Nickname), string.lower(name), 1, true)) then
            table.insert(results, cardData)
            addedCards[cardData.metadata.id] = true
      end
    end
  end
  return results
end

-- Gets a random basic weakness from the bag.  Once a given ID has been returned
-- it will be removed from the list and cannot be selected again until a reload
-- occurs or the indexes are rebuilt, which will refresh the list to include all
-- weaknesses.
-- Return: String ID of the selected weakness.
function getRandomWeaknessId()
    local availableWeaknesses = buildAvailableWeaknesses()
  if (#availableWeaknesses > 0) then
    return availableWeaknesses[math.random(#availableWeaknesses)]
  end
end

-- Constructs a list of available basic weaknesses by starting with the full pool of basic
-- weaknesses then removing any which are currently in the play or deck construction areas
-- Return: Table array of weakness IDs which are valid to choose from
function buildAvailableWeaknesses()
  local weaknessesInPlay = { }
  local allObjects = getAllObjects()
  for _, object in ipairs(allObjects) do
    if (object.name == "Deck" and isInPlayArea(object)) then
      for _, cardData in ipairs(object.getData().ContainedObjects) do
        local cardMetadata = JSON.decode(cardData.GMNotes)
        incrementWeaknessCount(weaknessesInPlay, cardMetadata)
      end
    elseif (object.name == "Card" and isInPlayArea(object)) then
      local cardMetadata = JSON.decode(object.getGMNotes())
      incrementWeaknessCount(weaknessesInPlay, cardMetadata)
    end
  end

  local availableWeaknesses = { }
  for _, weaknessId in ipairs(basicWeaknessList) do
    if (weaknessesInPlay[weaknessId] ~= nil and weaknessesInPlay[weaknessId] > 0) then
      weaknessesInPlay[weaknessId] = weaknessesInPlay[weaknessId] - 1
    else
      table.insert(availableWeaknesses, weaknessId)
    end
  end
  return availableWeaknesses
end

-- Helper function that adds one to the table entry for the number of weaknesses in play
function incrementWeaknessCount(table, cardMetadata)
  if (isBasicWeakness(cardMetadata)) then
    if (table[cardMetadata.id] == nil) then
      table[cardMetadata.id] = 1
    else
      table[cardMetadata.id] = table[cardMetadata.id] + 1
    end
  end
end

function isInPlayArea(object)
  if (object == nil) then
    return false
  end
  local position = object.getPosition()
  return position.x < WEAKNESS_CHECK_X
      and position.z < WEAKNESS_CHECK_Z
end
function isBasicWeakness(cardMetadata)
  return cardMetadata ~= nil
    and cardMetadata.weakness
    and cardMetadata.basicWeaknessCount ~= nil
    and cardMetadata.basicWeaknessCount > 0
end

end)
return __bundle_require("__root")`
	got, err := Unbundle(rawlua)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	want := `require("playercards/AllCardsBag")`
	if want != got {
		t.Errorf("want <%s>, got <%s>\n", want, got)
	}
}

func TestFailedUnbundle(t *testing.T) {
	rawlua := `  __bundle_register("core/AgendaDeck", function(require, _LOADED, __bundle_register, __bundle_modules)
	  MIN_VALUE = -99
	  MAX_VALUE = 999
`
	_, err := Unbundle(rawlua)
	if err == nil {
		t.Error("expected err, got no err")
	}
}

func TestNonBundled(t *testing.T) {
	rawlua := `
	  MIN_VALUE = -99
	  MAX_VALUE = 999
`
	got, err := Unbundle(rawlua)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}
	want := rawlua
	if want != got {
		t.Errorf("want <%s>, got <%s>\n", want, got)
	}
}
