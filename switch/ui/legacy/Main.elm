import Html exposing (..)
import Html.Attributes exposing (..)
import Http
import Json.Decode as Json exposing ((:=),string,at)
import Task
import String
import Navigation
import Dict


main : Program Never
main =
  Navigation.program urlParser
    { init = init
    , view = view
    , update = update
    , urlUpdate = urlUpdate
    , subscriptions = subscriptions
    }



-- URL PARSER


-- TODO: parameterize local vs public mode somehow

-- toUrl : Entity -> String
-- toUrl entity =
--   "#" ++ entity.address ++ toString entity.puerto


toUrl : Entity -> String
toUrl entity =
  "#https://localhost:" ++ toString entity.puerto


fromUrl : String -> Route
fromUrl url =
  Route (String.dropLeft 1 url) Overview


urlParser : Navigation.Parser Route
urlParser =
  Navigation.makeParser (fromUrl << .hash)

nodeBaseUrl : Entity -> String
nodeBaseUrl entity =
    case String.uncons entity.address of
        Nothing -> "/"
        Just a -> "https://" ++ entity.address ++ ":" ++ toString entity.puerto ++ "/"
apiUrl : Entity -> String
apiUrl entity =
  (nodeBaseUrl entity) ++ "swagger-ui"


-- MODEL


type ViewType
    = Overview
    | JsonView

type alias Route =
    { nodeAddress : String
    , viewType : ViewType
    }

type alias Model =
  { route : Route
  , health : Maybe Health
  , healthJson : String
  , error : Maybe String
  , loading: Bool
  }


init : Route -> (Model, Cmd Msg)
init route =
    urlUpdate route (Model route Nothing "nothing yet" Nothing False)



-- UPDATE

type Msg
  = UpdateRoute Route
  | FetchSucceed String
  | FetchFail Http.Error


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
  case msg of
    UpdateRoute route ->
        ({model | route = route}, Cmd.none )
    FetchSucceed healthString ->
      ({model | healthJson = healthString, health = Result.toMaybe (Json.decodeString decodeHealth healthString ), loading = False} , Cmd.none)

    FetchFail _ ->
      ({model | loading = False}, Cmd.none)


{-| The URL is turned into a result. If the URL is valid, we just update our
model to the new count. If it is not a valid URL, we modify the URL to make
sense.
-}
urlUpdate : Route -> Model -> (Model, Cmd Msg)
urlUpdate route model =
    ({model | route = route, loading = True}, getHealth route.nodeAddress)



-- VIEW


healthView : Health -> Html Msg
healthView health =
    div []
        [ h2 [] [ text health.entity.id ]
        , small []
            [ a [href (apiUrl health.entity)] [ text "API"] ]
        , h3 [] [ text "slaves" ]
        , entitiesUl health.slaves
        , h3 [] [ text "devices" ]
        , entitiesUl health.devices
        ]

entityLink : Entity -> Html Msg
entityLink entity =
  li [] [ a [href (toUrl entity)] [ text entity.id]]

entitiesUl : List Entity -> Html Msg
entitiesUl entities =
    ul [] (List.map entityLink entities)

view : Model -> Html Msg
view model =
  div []
      ( [ case model.health of
          Nothing ->
            div [] [ text "Nothing yet"]
          Just health -> healthView health
        , h3 [] [ text "JSON"]
        , pre [] [ text model.healthJson]
        ] ++ case model.error of
                Nothing -> []
                Just err -> [ text ("Error" ++ err) ]
      )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
  Sub.none



-- HTTP

getHealth : String -> Cmd Msg
getHealth nodeAddress =
  let
    url =
      nodeAddress ++ "/v1/health"
  in
    Task.perform FetchFail FetchSucceed (Http.get decodeJson url)

type alias Entity =
    { id : String
    , kind : String
    , address : String
    , puerto : Int
  -- FirstSeen time.Time
  -- LastSeen  time.Time
    }

type alias Health =
    { entity : Entity
    , slaves : List Entity
    , devices : List Entity
    }

decodeJson : Json.Decoder String
decodeJson = ("value" := string)
decodeEntity : Json.Decoder Entity
decodeEntity =
    Json.object4 Entity
        ("Id" := string)
        ("Kind" := string)
        ("Address" := string)
        ("Port" := Json.int)
dictValues : Json.Decoder b -> Json.Decoder (List b)
dictValues dec =
    Json.map Dict.values (Json.dict dec)
decodeHealth : Json.Decoder Health
decodeHealth =
    Json.object3 Health
        ( "Entity" := decodeEntity)
        ( "Slaves" := dictValues decodeEntity)
        ( "Devices" := dictValues decodeEntity)
